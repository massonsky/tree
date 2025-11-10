package config

import (
	"os"
	"path/filepath"
	"tree/pkg/assets"

	"gopkg.in/yaml.v3"
)

// Config структура конфигурационного файла
type Config struct {
	// Пути по умолчанию
	DefaultFontPath string `yaml:"default_font_path"`
	LogLevel        string `yaml:"log_level"`

	// Параметры генерации изображений
	ImageWidth  int `yaml:"image_width"`
	ImageHeight int `yaml:"image_height"`

	// CLI-настройки
	ShowHiddenFiles bool `yaml:"show_hidden_files"`
	MaxDepth        int  `yaml:"max_depth"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		DefaultFontPath: filepath.Join(GetConfigDir(), "fonts", "LiberationMono-Regular.ttf"),
		LogLevel:        "info",
		ImageWidth:      1200,
		ImageHeight:     0, // 0 = автоматический расчет
		ShowHiddenFiles: false,
		MaxDepth:        10,
	}
}

// EnsureConfig создает конфиг и директории если их нет
func EnsureConfig() (*Config, error) {
	configDir := GetConfigDir()
	logsDir := GetLogsDir()
	assetsDir := GetAssetsDir()
	fontsDir := GetFontsDir()
	colorSchemasDir := GetColorSchemasDir()
	templateImagesDir := GetTemplateImagesDir()
	configFile := GetConfigFile()

	// Создаем все необходимые директории
	dirs := []string{
		configDir,
		logsDir,
		assetsDir,
		fontsDir,
		colorSchemasDir,
		templateImagesDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	// Инициализируем стандартные файлы
	if err := ensureDefaultFiles(fontsDir, colorSchemasDir); err != nil {
		return nil, err
	}
	// Создаем конфиг если отсутствует
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := DefaultConfig()
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return nil, err
		}
	}

	// Читаем конфиг
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ensureDefaultFiles(fontsDir, colorSchemasDir string) error {
	// 1. Шрифт по умолчанию
	fontPath := filepath.Join(fontsDir, "Roboto-Black.ttf")
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		if err := os.WriteFile(fontPath, assets.DefaultFont, 0644); err != nil {
			return err
		}
	}

	// 2. Цветовая схема по умолчанию
	schemaPath := filepath.Join(colorSchemasDir, "default.yaml")
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		if err := os.WriteFile(schemaPath, assets.DefaultColorSchema, 0644); err != nil {
			return err
		}
	}

	return nil
}
