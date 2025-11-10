package exporter

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"tree/internal/types"

	svg "github.com/ajstarks/svgo"
)

type SVGExporter struct{}

func (e *SVGExporter) Export(w io.Writer, entries []types.Entry) error {
	if len(entries) == 0 {
		return fmt.Errorf("no entries to export")
	}

	height := calculateSVGHeight(entries)
	canvas := svg.New(w)
	canvas.Start(svgWidth, height)

	// Фон
	canvas.Rect(0, 0, svgWidth, height, "fill:#ffffff")

	// Стили текста
	fontStyle := fmt.Sprintf("font-family:monospace;font-size:%dpx", fontSize)

	y := padding + fontSize
	for i, entry := range entries {
		isLast := (i == len(entries)-1)
		prefix := buildTreePrefix(entry.Depth, isLast)

		name := filepath.Base(entry.Path)
		if entry.Info.IsDir() {
			name += "/"
		}

		line := prefix + name

		// Цвет текста: синий для директорий, чёрный для файлов
		color := "#000000"
		if entry.Info.IsDir() {
			color = "#1e88e5"
		}

		canvas.Text(padding, y, line, fontStyle+" fill:"+color)
		y += lineHeight
	}

	canvas.End()
	return nil
}

func calculateSVGHeight(entries []types.Entry) int {
	return padding*2 + (len(entries) * lineHeight)
}

// buildTreePrefix формирует префикс для дерева (как в console.go)
func buildTreePrefix(depth int, isLast bool) string {
	if depth <= 0 {
		return ""
	}

	parts := make([]string, depth)
	for d := 1; d < depth; d++ {
		parts[d-1] = "│   "
	}

	if isLast {
		parts[depth-1] = "└── "
	} else {
		parts[depth-1] = "├── "
	}

	return strings.Join(parts, "")
}
