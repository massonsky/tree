package exporter

import (
	"fmt"
	"io"
	"path/filepath"

	"tree/assets"
	"tree/internal/types"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

const (
	imageWidth = 1200
	lineHeight = 22
	padding    = 20
	fontSize   = 16
)

type PNGExporter struct {
	fontPath string
}

func NewPNGExporter(cfg map[string]interface{}) (Exporter, error) {
	fontPath, ok := cfg["font_path"].(string)
	if !ok {
		fontPath = ""
	}
	return &PNGExporter{fontPath: fontPath}, nil
}

func (e *PNGExporter) Export(w io.Writer, entries []types.Entry) error {
	if len(entries) == 0 {
		return fmt.Errorf("no entries to export")
	}

	height := calculateImageHeight(entries)
	dc := gg.NewContext(imageWidth, height)

	// Фон
	dc.SetRGB(0.98, 0.98, 0.98)
	dc.Clear()

	// Загрузка шрифта
	var err error
	if e.fontPath != "" {
		err = dc.LoadFontFace(e.fontPath, fontSize)
		if err != nil {
			err = loadFontFromBytes(dc, assets.DefaultFont, fontSize)
		}
	} else {
		err = loadFontFromBytes(dc, assets.DefaultFont, fontSize)
	}
	if err != nil {
		return fmt.Errorf("failed to load font: %w", err)
	}

	dc.SetRGB(0.1, 0.1, 0.1)
	y := padding + fontSize

	// Рисуем корень
	root := entries[0]
	rootName := root.Path
	if root.Info.IsDir() {
		rootName += "/"
	}
	dc.DrawString(rootName, padding, float64(y))
	y += lineHeight

	// Создаём массив для отслеживания, где нужны вертикальные линии
	// Для каждого уровня глубины: true = нужна вертикальная линия
	needVerticalLine := make([]bool, 20) // максимум 20 уровней

	for i := 1; i < len(entries); i++ {
		entry := entries[i]

		// Определяем, является ли текущий элемент последним в своей директории
		isLast := false
		if i+1 < len(entries) {
			nextEntry := entries[i+1]
			if nextEntry.Depth <= entry.Depth {
				isLast = true
			}
		} else {
			isLast = true
		}

		// Обновляем needVerticalLine для текущего уровня
		if entry.Depth > 0 {
			// Если текущий элемент НЕ последний — на этом уровне нужна вертикальная линия
			needVerticalLine[entry.Depth] = !isLast
		}

		// Формируем префикс
		prefix := ""
		for d := 1; d < entry.Depth; d++ {
			if needVerticalLine[d] {
				prefix += "|   "
			} else {
				prefix += "-    "
			}
		}

		if entry.Depth > 0 {
			if isLast {
				prefix += "`--"
			} else {
				prefix += "+-- "
			}
		}

		name := filepath.Base(entry.Path)
		if entry.Info.IsDir() {
			name += "/"
		}

		line := prefix + name
		dc.DrawString(line, padding, float64(y))
		y += lineHeight
	}

	return dc.EncodePNG(w)
}

func calculateImageHeight(entries []types.Entry) int {
	return padding*2 + (len(entries) * lineHeight)
}

// loadFontFromBytes загружает шрифт из []byte и устанавливает его в gg.Context
func loadFontFromBytes(dc *gg.Context, fontBytes []byte, size float64) error {
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("parse embedded font: %w", err)
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: size,
		DPI:  72,
	})
	dc.SetFontFace(face)
	return nil
}
