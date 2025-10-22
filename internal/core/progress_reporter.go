package core

import "github.com/evaneliasyoung/phaser-unpacker/internal/phaser"

type ProgressReporter interface {
	TextureProcessed(sheet phaser.Sheet, texture phaser.Texture)
	SheetProcessed(sheet phaser.Sheet)
	AllProcessed()
}

type NoopProgressReporter struct{}

func (*NoopProgressReporter) TextureProcessed(phaser.Sheet, phaser.Texture) {}
func (*NoopProgressReporter) SheetProcessed(phaser.Sheet)                   {}
func (*NoopProgressReporter) AllProcessed()                                 {}
