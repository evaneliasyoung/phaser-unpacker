package phaser

import "image"

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
