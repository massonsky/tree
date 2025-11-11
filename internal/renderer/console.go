package renderer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/massonsky/gotree/internal/config"
	"github.com/massonsky/gotree/internal/logger"
	_metrics "github.com/massonsky/gotree/internal/metrics"
	_type "github.com/massonsky/gotree/internal/types"
	"github.com/massonsky/gotree/internal/ui"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// PrintTree выводит структуру директории в консоль
func PrintTree(entries []_type.Entry, cfg *config.Config) {
	PrintTreeToWriter(os.Stdout, entries, cfg)
}

// PrintTreeToWriter выводит структуру директории в указанный writer (например, stdin pager'а).
func PrintTreeToWriter(w io.Writer, entries []_type.Entry, cfg *config.Config) {
	logger.Debugf("Rendering tree with %d entries", len(entries))

	if len(entries) == 0 {
		color.New(color.FgRed).Fprintln(w, "No files or directories found")
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
		printEntryToWriter(w, entry, isLast, width, maxDepth)
	}

	if cfg.LogLevel == "debug" {
		color.New(color.FgYellow).Fprintln(w, "Debug mode: showing hidden files")
		logger.Debug("Debug mode enabled")
	}
}
func shouldUseColor(mode string) bool {
	switch mode {
	case "never":
		return false
	case "always":
		return true
	case "auto":
		return term.IsTerminal(int(os.Stdout.Fd()))
	default:
		return term.IsTerminal(int(os.Stdout.Fd()))
	}
}

// printEntry выводит один элемент дерева с отступами
func printEntryToWriter(w io.Writer, entry _type.Entry, isLast bool, width int, maxDepth int) {
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
	style.Fprintln(w, line)
	// В конце функции добавляем логирование
	logger.Tracef("Rendered entry: %s (depth: %d, size: %d)",
		entry.Path, entry.Depth, entry.Info.Size())
}

func termSize() (int, int, error) {
	if ui.IsTerminal() {
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && width > 0 {
			return width, height, nil
		}
	}
	return 80, 24, nil
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

// PrintMetrics выводит собранные метрики
func PrintMetrics(m _metrics.Metrics) {
	fmt.Println()

	header := color.New(color.FgHiCyan, color.Bold).Sprint("📊 Scan Metrics")
	fmt.Println(header)

	fmt.Printf("   Files:       %s\n", color.GreenString("%d", m.TotalFiles))
	fmt.Printf("   Directories: %s\n", color.BlueString("%d", m.TotalDirs))
	fmt.Printf("   Total Size:  %s\n", color.YellowString("%s", _metrics.FormatSize(m.TotalSize)))
	fmt.Printf("   Max Depth:   %s\n", color.MagentaString("%d", m.MaxDepth))
	// форматируем длительность с большей точностью для очень коротких измерений
	var durationStr string
	if m.ScanDuration < time.Millisecond {
		durationStr = m.ScanDuration.String()
	} else {
		durationStr = m.ScanDuration.Truncate(time.Millisecond).String()
	}
	fmt.Printf("   Duration:    %s\n", color.WhiteString("%s", durationStr))

	// если скан был очень быстрым, не показываем вводящую в заблуждение скорость
	if m.ScanDuration < 10*time.Millisecond {
		fmt.Printf("   Performance: %s\n", color.CyanString("%s", "N/A (unstable, short duration)"))
	} else if m.FilesPerSecond > 0 {
		fmt.Printf("   Performance: %s\n", color.CyanString("%.1f files/sec", m.FilesPerSecond))
	}
}
