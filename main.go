package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	_ "golang.org/x/image/webp"
	"golang.org/x/term"
)

type Size struct {
	Width  int `json:"w"`
	Height int `json:"h"`
}

func (sz Size) Min() image.Point {
	return image.Point{0, 0}
}

func (sz Size) Max() image.Point {
	return image.Point{sz.Width, sz.Height}
}

func (sz Size) Rect() image.Rectangle {
	return image.Rectangle{sz.Min(), sz.Max()}
}

type Frame struct {
	Width  int `json:"w"`
	Height int `json:"h"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

func (fr Frame) Min() image.Point {
	return image.Point{fr.X, fr.Y}
}

func (fr Frame) Max() image.Point {
	return image.Point{fr.X + fr.Width, fr.Y + fr.Height}
}

func (fr Frame) Rect() image.Rectangle {
	return image.Rectangle{fr.Min(), fr.Max()}
}

type Texture struct {
	FileName         string `json:"filename"`
	Frame            Frame  `json:"frame"`
	Rotated          bool   `json:"rotated"`
	SourceSize       Size   `json:"sourceSize"`
	SpriteSourceSize Frame  `json:"spriteSourceSize"`
	Trimmed          bool   `json:"trimmed"`
}

type Sheet struct {
	Format   string    `json:"format"`
	Textures []Texture `json:"frames"`
	Image    string    `json:"image"`
	Scale    float64   `json:"scale"`
	Size     Size      `json:"size"`
}

type Pack struct {
	Meta   map[string]string `json:"meta"`
	Sheets []Sheet           `json:"textures"`
}

type Unpacker struct {
	Pack
	PackName  string
	InputDir  string
	OutputDir string
	Workers   int
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func (unpacker Unpacker) unpackTexture(texture Texture, img image.Image) error {
	spriteSize := texture.SourceSize.Rect()
	sprite := image.NewRGBA(spriteSize)

	destFrame := texture.SpriteSourceSize.Rect()
	sourceFrame := texture.Frame.Rect()

	draw.Draw(sprite, destFrame, img, sourceFrame.Min, draw.Src)

	encoder := png.Encoder{CompressionLevel: png.DefaultCompression}

	outputPath := filepath.Join(unpacker.OutputDir, texture.FileName+".png")

	if strings.Contains(texture.FileName, "/") {
		parts := strings.Split(texture.FileName, "/")
		subDir := filepath.Join(unpacker.OutputDir, filepath.Join(parts...))

		if err := os.MkdirAll(subDir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}

	if err = encoder.Encode(outputFile, sprite); err != nil {
		return fmt.Errorf("failed to encode sprite as png: %w", err)
	}

	if err = outputFile.Close(); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func (unpacker Unpacker) unpackSheet(sheet Sheet, sheetBar, totalBar *mpb.Bar) error {
	sheetPath := filepath.Join(unpacker.InputDir, sheet.Image)

	sheetFile, err := os.Open(sheetPath)
	if err != nil {
		return fmt.Errorf("failed to open texture sheet: %w", err)
	}

	img, _, err := image.Decode(sheetFile)
	if err != nil {
		return fmt.Errorf("failed to decode webp file: %w", err)
	}

	jobs := make(chan Texture)
	results := make(chan error, len(sheet.Textures))

	var wg sync.WaitGroup

	for range unpacker.Workers {
		wg.Go(func() {
			for tex := range jobs {
				if err := unpacker.unpackTexture(tex, img); err != nil {
					results <- err
					return
				}
				if sheetBar != nil {
					sheetBar.Increment()
				}
				if totalBar != nil {
					totalBar.Increment()
				}
				results <- nil
			}
		})
	}

	go func() {
		for _, tex := range sheet.Textures {
			jobs <- tex
		}
		close(jobs)
	}()

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

func (unpacker Unpacker) unpack(noProgress bool) error {
	numSheets := len(unpacker.Pack.Sheets)

	fmt.Printf("[info] found %d texture sheets\n", numSheets)
	fmt.Printf("[info] writing to %s\n", unpacker.OutputDir)

	if err := os.MkdirAll(unpacker.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var p *mpb.Progress = nil
	var sheetBars map[string]*mpb.Bar
	var totalBar *mpb.Bar

	totalTextures := 0

	if isTTY() && !noProgress {
		p = mpb.New()
		sheetBars = make(map[string]*mpb.Bar)
	}

	for _, sh := range unpacker.Sheets {
		totalTextures += len(sh.Textures)

		if p != nil {
			sheetBars[sh.Image] = p.AddBar(
				int64(len(sh.Textures)),
				mpb.PrependDecorators(
					decor.Name(sh.Image+" ", decor.WCSyncWidth),
					decor.CountersNoUnit("%d / %d"),
				),
				mpb.AppendDecorators(
					decor.Percentage(),
				),
			)
		}
	}

	if p != nil {
		totalBar = p.AddBar(
			int64(totalTextures),
			mpb.PrependDecorators(
				decor.Name("Total ", decor.WCSyncWidth),
				decor.CountersNoUnit("%d / %d"),
			),
			mpb.AppendDecorators(
				decor.Percentage(),
			),
		)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, sh := range unpacker.Sheets {
		wg.Add(1)

		go func(sh Sheet, sBar, tBar *mpb.Bar) {
			defer wg.Done()
			if err := unpacker.unpackSheet(sh, sBar, tBar); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(sh, sheetBars[sh.Image], totalBar)
	}

	wg.Wait()
	if p != nil {
		p.Wait()
	}

	if firstErr != nil {
		return firstErr
	}

	fmt.Printf("[info] extracted %d textures from %d sheets\n", totalTextures, len(unpacker.Sheets))

	return nil
}

func main() {
	var outputDir string
	var workers int = 2 * runtime.NumCPU()
	var noProgress bool = false

	if workers > 32 {
		workers = 32
	}

	var rootCmd = &cobra.Command{
		Use:   "txunpak <path>",
		Short: "Unpoack Phaser assets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var path = args[0]

			if filepath.Ext(path) != ".json" {
				return fmt.Errorf("input file must be a .json file")
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			var pack Pack
			if err := json.Unmarshal(data, &pack); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			inputDir := filepath.Dir(path)
			packName := strings.TrimSuffix(inputDir, ".json")
			if outputDir == "" {
				outputDir = filepath.Join(filepath.Dir(path), packName)
			}

			unpacker := Unpacker{
				Pack:      pack,
				PackName:  packName,
				InputDir:  inputDir,
				OutputDir: outputDir,
				Workers:   workers,
			}

			return unpacker.unpack(noProgress)
		},
	}

	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", workers, "Number of concurrent workers")
	rootCmd.Flags().BoolVarP(&noProgress, "no-progress", "", noProgress, "Disable progress bars")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
