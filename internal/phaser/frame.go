package phaser

import "image"

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
