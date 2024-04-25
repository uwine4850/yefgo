package codegen

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Arg struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Method struct {
	Type   string `yaml:"type"`
	Args   []Arg  `yaml:"args"`
	Output string `yaml:"output"`
}

type Class struct {
	Name    string            `yaml:"name"`
	Args    []Arg             `yaml:"args"`
	Methods map[string]Method `yaml:"methods"`
}

type Function struct {
	Name   string `yaml:"name"`
	Args   []Arg  `yaml:"args"`
	Output string `yaml:"output"`
}

type Module struct {
	Classes   []Class    `yaml:"classes"`
	Functions []Function `yaml:"functions"`
	GoPkg     string     `yaml:"gopkg"`
	FileName  string     `yaml:"filename"`
	Import    []string   `yaml:"import"`
}

type Config struct {
	Modules map[string]Module `yaml:"modules"`
}

func ParseYefCfg(filename string) (Config, error) {
	var config Config
	file, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
