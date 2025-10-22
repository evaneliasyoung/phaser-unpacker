package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/evaneliasyoung/phaser-unpacker/internal/core"
	"github.com/evaneliasyoung/phaser-unpacker/internal/phaser"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func unpack(unpacker core.Unpacker, noProgress bool) error {
	numSheets := len(unpacker.Pack.Sheets)

	fmt.Printf("[info] found %d texture sheets\n", numSheets)
	fmt.Printf("[info] writing to %s\n", unpacker.OutputDir)

	totalTextures := 0
	reporter := makeReporter(unpacker, &totalTextures, noProgress)

	if err := unpacker.UnpackAll(reporter); err != nil {
		return err
	}

	fmt.Printf("[info] extracted %d textures from %d sheets\n", totalTextures, len(unpacker.Sheets))

	return nil
}

func Parse() {
	var outputDir string
	var workers int = 2 * runtime.NumCPU()
	var noProgress bool = false

	if workers > 32 {
		workers = 32
	}

	var rootCmd = &cobra.Command{
		Use:   "phaser-unpacker <path>",
		Short: "Unpack Phaser assets",
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

			var pack phaser.Pack
			if err := json.Unmarshal(data, &pack); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			inputDir := filepath.Dir(path)
			packName := strings.TrimSuffix(inputDir, ".json")
			if outputDir == "" {
				outputDir = filepath.Join(filepath.Dir(path), packName)
			}

			unpacker := core.Unpacker{
				Pack:      pack,
				PackName:  packName,
				InputDir:  inputDir,
				OutputDir: outputDir,
				Workers:   workers,
			}

			return unpack(unpacker, noProgress)
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
