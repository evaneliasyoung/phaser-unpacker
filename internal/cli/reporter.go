package cli

import (
	"github.com/evaneliasyoung/phaser-unpacker/internal/core"
	"github.com/evaneliasyoung/phaser-unpacker/internal/phaser"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type reporter struct {
	p      *mpb.Progress
	sheets map[string]*mpb.Bar
	total  *mpb.Bar
}

func (r *reporter) TextureProcessed(sh phaser.Sheet, _ phaser.Texture) {
	if r.p != nil {
		r.sheets[sh.Image].Increment()
		r.total.Increment()
	}
}

func (r *reporter) SheetProcessed(phaser.Sheet) {}

func (r *reporter) AllProcessed() {
	if r.p != nil {
		r.p.Wait()
	}
}

func makeReporter(unpacker core.Unpacker, totalTextures *int, noProgress bool) *reporter {
	if !(isTTY() && !noProgress) {
		return &reporter{p: nil, sheets: nil, total: nil}
	}

	p := mpb.New()
	sheets := make(map[string]*mpb.Bar)

	for _, sh := range unpacker.Sheets {
		*totalTextures += len(sh.Textures)

		sheets[sh.Image] = p.AddBar(
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

	total := p.AddBar(
		int64(*totalTextures),
		mpb.PrependDecorators(
			decor.Name("Total ", decor.WCSyncWidth),
			decor.CountersNoUnit("%d / %d"),
		),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)

	return &reporter{p: p, sheets: sheets, total: total}
}
