package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Template struct {
	Prefix struct {
		Vertical string `yaml:"vertical"`
		Corner   string `yaml:"corner"`
		Branch   string `yaml:"branch"`
	} `yaml:"prefix"`
	Icons struct {
		File string `yaml:"file"`
		Dir  string `yaml:"dir"`
	} `yaml:"icons"`
	Colors struct {
		File string `yaml:"file"`
		Dir  string `yaml:"dir"`
	} `yaml:"colors"`
}

func LoadTemplate(templatesDir, name string) (*Template, error) {
	if name == "" {
		name = "default"
	}

	path := filepath.Join(templatesDir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tpl Template
	if err := yaml.Unmarshal(data, &tpl); err != nil {
		return nil, err
	}

	// Устанавливаем значения по умолчанию
	if tpl.Prefix.Vertical == "" {
		tpl.Prefix.Vertical = "│"
	}
	if tpl.Prefix.Corner == "" {
		tpl.Prefix.Corner = "└──"
	}
	if tpl.Prefix.Branch == "" {
		tpl.Prefix.Branch = "├──"
	}
	if tpl.Colors.File == "" {
		tpl.Colors.File = "#000000"
	}
	if tpl.Colors.Dir == "" {
		tpl.Colors.Dir = "#1e88e5"
	}

	return &tpl, nil
}
