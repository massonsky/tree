package tree

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tree/internal/config"
	"tree/internal/logger"
	"tree/internal/metrics"
	"tree/internal/types"
	"tree/internal/ui"

	"github.com/gobwas/glob"
	"github.com/schollz/progressbar/v3"
)

// WalkDirWithContext обходит директорию с прогрессом в реальном времени
func WalkDirWithContext(
	ctx context.Context,
	root string,
	cfg *config.Config,
	progressEnabled bool,
) (WalkResult, error) {
	startTime := time.Now()
	logger.Debugf("Starting directory walk with progress: %t", progressEnabled)

	root, err := filepath.Abs(root)
	if err != nil {
		return WalkResult{}, err
	}

	// НЕ считаем общее количество файлов — экономим время!
	var bar *progressbar.ProgressBar
	if progressEnabled {
		// Indeterminate mode: max = -1 или 0
		bar = ui.NewProgressBar(-1, "Scanning files", ui.DefaultProgressBarConfig())
		defer bar.Finish()
		ctx = ui.WithCancel(ctx, bar)
	}

	var entries []types.Entry

	// Добавляем корневой элемент
	rootInfo, err := os.Stat(root)
	if err != nil {
		return WalkResult{}, err
	}
	entries = append(entries, types.Entry{
		Path:  filepath.Base(root),
		Info:  rootInfo,
		Depth: 0,
	})

	// Основной обход (один проход!)
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			logger.Warn("Directory walk cancelled by user")
			return ctx.Err()
		default:
		}

		if path == root {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		// Скрытые файлы
		if !cfg.ShowHiddenFiles && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Глубина
		relPath, _ := filepath.Rel(root, path)
		depth := len(strings.Split(relPath, string(filepath.Separator)))
		if cfg.MaxDepth > 0 && depth > cfg.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Игнорирование
		relPathForMatch := filepath.ToSlash(relPath)
		for _, pattern := range cfg.IgnorePatterns {
			g, err := glob.Compile(pattern)
			if err != nil {
				continue
			}
			if g.Match(relPathForMatch) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Добавляем запись
		entries = append(entries, types.Entry{
			Path:  relPath,
			Info:  info,
			Depth: depth,
		})

		// Обновляем прогресс в реальном времени
		if progressEnabled && bar != nil {
			bar.Add(1)
		}

		return nil
	})

	if err != nil {
		logger.Errorf("Directory walk failed: %v", err)
		return WalkResult{}, err
	}

	mets := metrics.Collect(entries, startTime)
	logger.Infof("Found %d entries in %s", len(entries)-1, root)

	return WalkResult{
		Entries: entries,
		Metrics: mets,
	}, nil
}

func WalkDir(root string, cfg *config.Config) ([]types.Entry, error) {
	result, err := WalkDirWithContext(context.Background(), root, cfg, true)
	if err != nil {
		return nil, err
	}
	return result.Entries, nil
}
