package exporter

import (
	"fmt"
	"io"
	_types "tree/internal/types"
)

// Exporter интерфейс для всех форматов экспорта
type Exporter interface {
	Export(w io.Writer, entries []_types.Entry) error
}

// Format поддерживаемые форматы
type Format string

const (
	FormatPNG  Format = "png"
	FormatTXT  Format = "txt"
	FormatJSON Format = "json"
	FormatSVG  Format = "svg" // Добавили SVG
)

// New создает экспортер по формату
func New(format Format, config map[string]interface{}) (Exporter, error) {
	switch format {
	case FormatPNG:
		return NewPNGExporter(config)
	case FormatTXT:
		return &TextExporter{}, nil
	case FormatJSON:
		return &JSONExporter{}, nil
	case FormatSVG:
		return &SVGExporter{}, nil
	default:
		return nil, ErrUnsupportedFormat
	}
}

// ErrUnsupportedFormat is returned when an unsupported export format is requested.
var ErrUnsupportedFormat = fmt.Errorf("unsupported export format")
