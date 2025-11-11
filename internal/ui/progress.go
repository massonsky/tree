package ui

import (
	"context"
	"os"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

// ProgressBarConfig настройки прогресс-бара
type ProgressBarConfig struct {
	EnableColors bool
	ShowBytes    bool
	ShowCount    bool
	ShowIts      bool
}

// DefaultProgressBarConfig возвращает настройки по умолчанию
func DefaultProgressBarConfig() ProgressBarConfig {
	return ProgressBarConfig{
		EnableColors: term.IsTerminal(int(os.Stdout.Fd())),
		ShowBytes:    true,
		ShowCount:    true,
		ShowIts:      true,
	}
}

// NewProgressBar создает настроенный прогресс-бар
func NewProgressBar(max int64, description string, cfg ProgressBarConfig) *progressbar.ProgressBar {
	options := []progressbar.Option{
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionOnCompletion(func() {
			if cfg.EnableColors {
				print("\033[32m✓\033[0m Done!\n")
			} else {
				println(" done")
			}
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	}
	if cfg.ShowBytes {
		options = append(options, progressbar.OptionShowBytes(true))
	}
	if cfg.EnableColors {
		options = append(options, progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "➤",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	}

	if cfg.ShowCount {
		options = append(options, progressbar.OptionShowCount())
	}

	if cfg.ShowIts {
		options = append(options, progressbar.OptionShowIts())
	}

	// Показываем ETA только если известно общее количество
	if max > 0 {
		options = append(options, progressbar.OptionSetPredictTime(true))
	}

	return progressbar.NewOptions64(max, options...)
}

// WithCancel добавляет обработку контекста для прогресс-бара
func WithCancel(ctx context.Context, bar *progressbar.ProgressBar) context.Context {
	// Канал для отслеживания прерывания
	stopCh := make(chan struct{})
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-ctx.Done():
			bar.Finish()
		case <-stopCh:
			cancel()
		}
	}()

	return context.WithValue(ctx, "progressStop", stopCh)
}

// StopProgressBar останавливает прогресс-бар и освобождает ресурсы
func StopProgressBar(ctx context.Context, bar *progressbar.ProgressBar) {
	if stopCh, ok := ctx.Value("progressStop").(chan struct{}); ok {
		close(stopCh)
	}
	bar.Finish()
}
