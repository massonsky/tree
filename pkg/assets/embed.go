package assets

import (
	_ "embed"
)

//go:embed fonts/Roboto-Black.ttf
var DefaultFont []byte

//go:embed color_schemas/default.yaml
var DefaultColorSchema []byte
