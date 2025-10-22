package core

import (
	"fmt"
	"os"
	"sync"

	"github.com/evaneliasyoung/phaser-unpacker/internal/phaser"
)

func (unpacker Unpacker) UnpackAll(reporter ProgressReporter) error {
	if err := os.MkdirAll(unpacker.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, sh := range unpacker.Sheets {
		wg.Add(1)

		go func(sh phaser.Sheet, rp ProgressReporter) {
			defer wg.Done()
			if err := unpacker.UnpackSheet(sh, rp); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(sh, reporter)
	}

	wg.Wait()

	reporter.AllProcessed()

	if firstErr != nil {
		return firstErr
	}

	return nil
}
