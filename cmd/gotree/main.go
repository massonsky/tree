package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/massonsky/gotree/internal/config"
	"github.com/massonsky/gotree/internal/exporter"
	"github.com/massonsky/gotree/internal/logger"
	"github.com/massonsky/gotree/internal/renderer"
	"github.com/massonsky/gotree/internal/tree"
	"github.com/massonsky/gotree/internal/tui"

	"github.com/urfave/cli/v2"
)

var appConfig *config.Config

func getFormatFromExtension(filename string) exporter.Format {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return exporter.FormatPNG
	case ".txt":
		return exporter.FormatTXT
	case ".json":
		return exporter.FormatJSON
	case ".svg": // Добавили SVG
		return exporter.FormatSVG
	default:
		if strings.Contains(strings.ToLower(filename), "json") {
			return exporter.FormatJSON
		}
		return exporter.FormatTXT
	}
}

// parseIgnorePatternsFromSlice нормализует значения --ignore, поддерживает
// одиночные элементы, пробельное разделение и список в квадратных скобках
func parseIgnorePatternsFromSlice(raw []string) []string {
	var out []string
	for _, item := range raw {
		s := strings.TrimSpace(item)
		s = strings.TrimPrefix(s, "[")
		s = strings.TrimSuffix(s, "]")

		if strings.Contains(s, ",") {
			parts := strings.Split(s, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					out = append(out, p)
				}
			}
			continue
		}

		if strings.Contains(s, " ") {
			parts := strings.Fields(s)
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					out = append(out, p)
				}
			}
			continue
		}

		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// processDirectory — основная логика обработки директории
func processDirectory(ctx context.Context, c *cli.Context, path string) error {
	logger.Infof("Processing directory: %s", path)

	// Применяем флаги в конфиг ДО старта обхода
	if c.IsSet("depth") {
		appConfig.MaxDepth = c.Int("depth")
	}
	if c.IsSet("ignore") {
		appConfig.IgnorePatterns = parseIgnorePatternsFromSlice(c.StringSlice("ignore"))
	}

	showProgress := !c.Bool("no-progress")
	walkResult, err := tree.WalkDirWithContext(ctx, path, appConfig, showProgress)
	if err != nil {
		if err == context.Canceled {
			logger.Info("Operation cancelled by user")
			return nil
		}
		logger.Errorf("WalkDir failed: %v", err)
		return cli.Exit(err.Error(), 1)
	}
	if c.IsSet("ignore") {
		appConfig.IgnorePatterns = parseIgnorePatternsFromSlice(c.StringSlice("ignore"))
	}
	// ЭКСПОРТ В ФАЙЛ
	if exportPath := c.String("export"); exportPath != "" {
		format := getFormatFromExtension(exportPath)
		config := make(map[string]interface{})
		config["templates_dir"] = appConfig.TemplatesDir
		config["template"] = c.String("template")

		if c.String("font") != "" {
			config["font_path"] = c.String("font")
		}
		if fontPath := c.String("font"); fontPath != "" {
			config["font_path"] = fontPath
		}

		exporterImpl, err := exporter.New(format, config)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Export error: %v", err), 1)
		}

		file, err := os.Create(exportPath)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Cannot create file %s: %v", exportPath, err), 1)
		}
		defer file.Close()

		if err := exporterImpl.Export(file, walkResult.Entries); err != nil {
			return cli.Exit(fmt.Sprintf("Export failed: %v", err), 1)
		}
		if !c.Bool("no-metrics") {
			renderer.PrintMetrics(walkResult.Metrics)
		}
		logger.Infof("Exported to %s", exportPath)
		return nil
	}

	// ОБЫЧНЫЙ ВЫВОД В КОНСОЛЬ
	renderer.PrintTree(walkResult.Entries, appConfig)

	if !c.Bool("no-metrics") {
		renderer.PrintMetrics(walkResult.Metrics)
	}

	logger.Infof("Successfully rendered tree for %s", path)
	return nil
}

func main() {
	// Загружаем конфиг
	var err error
	appConfig, err = config.EnsureConfig()
	if err != nil {
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

	// Обработка Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Общие флаги для всех команд
	commonFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "export",
			Aliases: []string{"e"},
			Usage:   "Export tree to file (supports: png, txt, json)",
		},
		&cli.StringFlag{
			Name:  "font",
			Usage: "Path to TTF font file for PNG export",
		},
		&cli.BoolFlag{
			Name:    "no-progress",
			Aliases: []string{"np"},
			Usage:   "Disable progress bar",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "no-metrics",
			Aliases: []string{"nm"},
			Usage:   "Hide scan metrics",
			Value:   false,
		},
		&cli.IntFlag{
			Name:  "depth",
			Usage: "Max depth of directory tree",
			Value: 10, // значение по умолчанию
		},
		&cli.StringSliceFlag{
			Name:    "ignore",
			Aliases: []string{"I"},
			Usage:   "Ignore paths matching pattern (can be used multiple times)",
		},
	}

	app := &cli.App{
		Name:  "gotree",
		Usage: "📁 Advanced directory tree visualizer",
		Flags: commonFlags,
		Action: func(c *cli.Context) error {
			path := "."
			if c.Args().Present() {
				path = c.Args().First()
			}
			return processDirectory(ctx, c, path)
		},
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "manage configuration",
				Subcommands: []*cli.Command{
					{
						Name:  "edit",
						Usage: "edit configuration in $EDITOR",
						Action: func(c *cli.Context) error {
							newCfg, err := config.EditConfigInteractive()
							if err != nil {
								logger.Errorf("Config edit failed: %v", err)
								return cli.Exit(err.Error(), 1)
							}
							logger.Infof("Config updated. New default font: %s", newCfg.DefaultFontPath)
							return nil
						},
					},
				},
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "render a directory tree",
				Flags:   commonFlags,
				Action: func(c *cli.Context) error {
					path := "."
					if c.Args().Present() {
						path = c.Args().First()
					}
					return processDirectory(ctx, c, path)
				},
			},
			{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "interactive tree explorer",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "no-progress",
						Aliases: []string{"np"},
						Usage:   "Disable progress bar during initial scan",
						Value:   false,
					},
				},
				Action: func(c *cli.Context) error {
					path := "."
					if c.Args().Present() {
						path = c.Args().First()
					}

					// Обновляем MaxDepth для интерактивного режима (больше глубины)
					appConfig.MaxDepth = 20

					logger.Infof("Starting interactive mode for %s", path)
					return tui.Run(ctx, appConfig, path)
				},
			},
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		logger.Errorf("Application failed: %v", err)
		os.Exit(1)
	}
}
