package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"tree/assets"

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
	ShowHiddenFiles bool     `yaml:"show_hidden_files"`
	MaxDepth        int      `yaml:"max_depth"`
	IgnorePatterns  []string `yaml:"ignore_patterns"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		DefaultFontPath: filepath.Join(GetConfigDir(), "fonts", "Roboto-Black.ttf"),
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

// UpdateConfig сохраняет переданную конфигурацию в файл конфигурации
func UpdateConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(GetConfigFile(), data, 0644)
}

// EditConfigInteractive открывает конфиг в внешнем редакторе (ENV $EDITOR) и
// сохраняет изменения обратно в файл конфигурации после валидации YAML.
// Возвращает обновлённую конфигурацию или ошибку.
func EditConfigInteractive() (*Config, error) {
	// Убедимся, что конфиг и директории созданы
	if _, err := EnsureConfig(); err != nil {
		return nil, err
	}

	// Прочитаем текущий конфиг
	data, err := os.ReadFile(GetConfigFile())
	if err != nil {
		return nil, err
	}

	// Создадим временный файл с текущим конфигом
	tmp, err := os.CreateTemp("", "tree-config-*.yaml")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return nil, err
	}
	_ = tmp.Close()

	// Выберем редактор
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi"
		}
	}

	// Откроем редактор и дождёмся его завершения
	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}

	// Прочитаем изменённый файл
	newData, err := os.ReadFile(tmpPath)
	if err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}

	// Валидация YAML
	var newCfg Config
	if err := yaml.Unmarshal(newData, &newCfg); err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}

	// Сохраним конфиг
	if err := UpdateConfig(&newCfg); err != nil {
		_ = os.Remove(tmpPath)
		return nil, err
	}

	_ = os.Remove(tmpPath)
	return &newCfg, nil
}
