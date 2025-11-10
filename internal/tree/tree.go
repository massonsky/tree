package tree

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"tree/internal/config"
	"tree/internal/logger"
	// Новый импорт
)

// Entry представляет элемент файловой системы
type Entry struct {
	Path  string
	Info  os.FileInfo
	Depth int // Глубина вложенности для форматирования вывода
}

// WalkDir обходит директорию и возвращает структурированные данные
func WalkDir(root string, cfg *config.Config) ([]Entry, error) {
	logger.Debugf("Starting directory walk: %s", root)

	var entries []Entry

	// Приводим путь к абсолютному
	root, err := filepath.Abs(root)
	if err != nil {
		logger.Errorf("Failed to get absolute path for %s: %v", root, err)
		return nil, err
	}

	// Получаем информацию о корневой директории
	rootInfo, err := os.Stat(root)
	if err != nil {
		logger.Errorf("Failed to stat root directory %s: %v", root, err)
		return nil, err
	}

	entries = append(entries, Entry{
		Path:  filepath.Base(root),
		Info:  rootInfo,
		Depth: 0,
	})

	err = filepath.WalkDir(
		root,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				logger.Warnf("Error accessing %s: %v", path, err)
				return nil // Продолжаем обход несмотря на ошибку
			}

			// Пропускаем корневой элемент
			if path == root {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				logger.Warnf("Failed to get info for %s: %v", path, err)
				return nil
			}

			// Применяем настройки из конфига
			if !cfg.ShowHiddenFiles && strings.HasPrefix(d.Name(), ".") {
				if d.IsDir() {
					logger.Debugf("Skipping hidden directory: %s", path)
					return filepath.SkipDir
				}
				logger.Debugf("Skipping hidden file: %s", path)
				return nil
			}

			// Вычисляем глубину вложенности
			relPath, _ := filepath.Rel(root, path)
			depth := len(strings.Split(relPath, string(filepath.Separator)))

			// Проверка глубины
			if cfg.MaxDepth > 0 && depth > cfg.MaxDepth {
				if d.IsDir() {
					logger.Debugf("Skipping directory beyond max depth (%d): %s", cfg.MaxDepth, path)
					return filepath.SkipDir
				}
				logger.Debugf("Skipping file beyond max depth (%d): %s", cfg.MaxDepth, path)
				return nil
			}

			entries = append(entries, Entry{
				Path:  relPath,
				Info:  info,
				Depth: depth,
			})

			return nil
		},
	)

	if err != nil {
		logger.Errorf("Directory walk failed: %v", err)
		return nil, err
	}

	logger.Infof("Found %d entries in %s", len(entries), root)
	return entries, nil
}
