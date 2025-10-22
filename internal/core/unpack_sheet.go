package core

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sync"

	"github.com/evaneliasyoung/phaser-unpacker/internal/phaser"
)

func (unpacker Unpacker) UnpackSheet(sheet phaser.Sheet, reporter ProgressReporter) error {
	sheetPath := filepath.Join(unpacker.InputDir, sheet.Image)

	sheetFile, err := os.Open(sheetPath)
	if err != nil {
		return fmt.Errorf("failed to open texture sheet: %w", err)
	}

	img, _, err := image.Decode(sheetFile)
	if err != nil {
		return fmt.Errorf("failed to decode webp file: %w", err)
	}

	jobs := make(chan phaser.Texture)
	results := make(chan error, len(sheet.Textures))

	var wg sync.WaitGroup

	for range unpacker.Workers {
		wg.Go(func() {
			for tex := range jobs {
				if err := unpacker.UnpackTexture(tex, img); err != nil {
					results <- err
					return
				}
				reporter.TextureProcessed(sheet, tex)
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

	reporter.SheetProcessed(sheet)

	return nil
}
