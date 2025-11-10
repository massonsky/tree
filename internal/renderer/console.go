package renderer

import (
	"fmt"
	"os"
	"path/filepath"

	"tree/internal/config"
	"tree/internal/logger"
	"tree/internal/tree"

	"github.com/fatih/color"
)

// PrintTree выводит структуру директории в консоль
func PrintTree(entries []tree.Entry, cfg *config.Config) {
	logger.Debugf("Rendering tree with %d entries", len(entries))

	if len(entries) == 0 {
		color.Red("No files or directories found")
		logger.Warn("No entries to render")
		return
	}

	width, _, _ := termSize()
	maxDepth := 0

	for _, entry := range entries {
		if entry.Depth > maxDepth {
			maxDepth = entry.Depth
		}
	}

	logger.Debugf("Terminal width: %d, Max depth: %d", width, maxDepth)

	// Выводим каждый элемент
	for i, entry := range entries {
		isLast := (i == len(entries)-1)
		printEntry(entry, isLast, width, maxDepth)
	}

	if cfg.LogLevel == "debug" {
		color.Yellow("Debug mode: showing hidden files")
		logger.Debug("Debug mode enabled")
	}
}

// printEntry выводит один элемент дерева с отступами
func printEntry(entry tree.Entry, isLast bool, width int, maxDepth int) {
	// Формируем префикс для отступов
	prefix := ""
	if entry.Depth > 0 {
		for d := 1; d < entry.Depth; d++ {
			prefix += "│   "
		}
		if isLast {
			prefix += "└── "
		} else {
			prefix += "├── "
		}
	}

	// Определяем иконку и цвет
	icon := "📄"
	style := color.New(color.FgWhite)

	if entry.Info.IsDir() {
		icon = "📁"
		style = color.New(color.FgCyan, color.Bold)
	}

	// Обрезаем длинные имена под ширину терминала
	displayName := filepath.Base(entry.Path)
	if entry.Depth == 0 {
		displayName = entry.Path
	}

	maxNameLength := width - len(prefix) - 10 // 10 для иконки и буфера
	if len(displayName) > maxNameLength && maxNameLength > 10 {
		displayName = displayName[:maxNameLength-3] + "..."
	}

	// Формируем строку
	line := fmt.Sprintf("%s%s %s", prefix, icon, displayName)

	// Добавляем информацию о размере для файлов
	if !entry.Info.IsDir() {
		size := formatSize(entry.Info.Size())
		line += fmt.Sprintf(" (%s)", size)
	}

	// Выводим с цветовым выделением
	style.Println(line)
	// В конце функции добавляем логирование
	logger.Tracef("Rendered entry: %s (depth: %d, size: %d)",
		entry.Path, entry.Depth, entry.Info.Size())
}

// Вспомогательные функции
func termSize() (int, int, error) {
	width, height, err := defaultTermSize()
	if err != nil || width == 0 {
		return 80, 24, nil // Значения по умолчанию
	}
	return width, height, err
}

func defaultTermSize() (int, int, error) {
	// Простая реализация для кросс-платформенности
	_, err := os.Stdout.Stat()
	if err != nil {
		return 0, 0, err
	}

	// Для Unix-систем используем ioctl (упрощенная версия)
	// В реальном проекте лучше использовать github.com/mattn/go-isatty или github.com/wayneashleyberry/terminal
	return 80, 24, nil
}

func formatSize(bytes int64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota)
		MB
		GB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
