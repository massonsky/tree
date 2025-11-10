package main

import (
	"log"
	"os"

	"tree/internal/config"
	"tree/internal/logger" // Новый импорт
	"tree/internal/renderer"
	"tree/internal/tree"

	"github.com/urfave/cli/v2"
)

var appConfig *config.Config

func main() {
	// Загружаем конфиг
	var err error
	appConfig, err = config.EnsureConfig()
	if err != nil {
		// Используем базовое логирование до инициализации
		log.Printf("FATAL: Config error: %v", err)
		os.Exit(1)
	}

	// Инициализируем логгер
	logDir := config.GetLogsDir()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("FATAL: Cannot create log directory: %v", err)
		os.Exit(1)
	}

	if err := logger.Init(appConfig); err != nil {
		log.Printf("FATAL: Logger init failed: %v", err)
		os.Exit(1)
	}

	logger.Info("Application started. Version: 1.0.0")
	defer logger.Info("Application terminated")

	app := &cli.App{
		Name:  "three",
		Usage: "📁 Advanced directory tree visualizer",
		Action: func(c *cli.Context) error {
			path := "."
			if c.Args().Present() {
				path = c.Args().First()
			}

			logger.Infof("Processing directory: %s", path)

			// Используем настройки из конфига
			entries, err := tree.WalkDir(path, appConfig)
			if err != nil {
				logger.Errorf("WalkDir failed: %v", err)
				return cli.Exit(err.Error(), 1)
			}

			logger.Debugf("Found %d entries", len(entries))
			renderer.PrintTree(entries, appConfig)

			logger.Infof("Successfully rendered tree for %s", path)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Errorf("Application failed: %v", err)
		os.Exit(1)
	}
}
