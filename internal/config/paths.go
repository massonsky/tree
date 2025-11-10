package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigDir возвращает путь к .tree конфигурации
func GetConfigDir() string {
	var configDir string

	// Определяем ОС
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\.tree
		appData := os.Getenv("APPDATA")
		if appData == "" {
			// Fallback для Windows
			home := os.Getenv("USERPROFILE")
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, ".tree")
	case "darwin":
		// macOS: ~/Library/Application Support/.tree
		home := os.Getenv("HOME")
		configDir = filepath.Join(home, "Library", "Application Support", ".tree")
	default:
		// Linux/BSD: ~/.config/.tree
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			home := os.Getenv("HOME")
			configHome = filepath.Join(home, ".config")
		}
		configDir = filepath.Join(configHome, ".tree")
	}

	return configDir
}

// GetLogsDir возвращает путь к логам
func GetLogsDir() string {
	return filepath.Join(GetConfigDir(), "log")
}

// GetConfigFile возвращает путь к YAML-конфигу
func GetConfigFile() string {
	return filepath.Join(GetConfigDir(), "configuration.yaml")
}

// GetAssetsDir возвращает путь к директории assets
func GetAssetsDir() string {
	return filepath.Join(GetConfigDir(), "assets")
}

// GetFontsDir возвращает путь к шрифтам
func GetFontsDir() string {
	return filepath.Join(GetAssetsDir(), "fonts")
}

// GetColorSchemasDir возвращает путь к цветовым схемам
func GetColorSchemasDir() string {
	return filepath.Join(GetAssetsDir(), "color_schemas")
}

// GetTemplateImagesDir возвращает путь к шаблонам изображений
func GetTemplateImagesDir() string {
	return filepath.Join(GetConfigDir(), "template_images")
}
