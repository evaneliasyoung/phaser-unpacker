package phaser

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
