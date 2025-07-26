package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	OutputDir  string `json:"outputDir"`
	InputDir   string `json:"inputDir"`
	StaticsDir string `json:"staticsDir"`
}

var cfg *Config

func Load(fp string) error {
	cfgFile, err := os.ReadFile(fp)
	if err != nil {
		return err
	}
	cfg = &Config{}
	return json.Unmarshal(cfgFile, cfg)
}

func Get() *Config {
	if cfg == nil {
		panic("Config not loaded, call Load() first")
	}
	return cfg
}
