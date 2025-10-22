package core

import "github.com/evaneliasyoung/phaser-unpacker/internal/phaser"

type Unpacker struct {
	phaser.Pack
	PackName  string
	InputDir  string
	OutputDir string
	Workers   int
}
