package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port            int    `json:"port"`
	BaseURL         string `json:"baseUrl"`
	ClientID        string `json:"clientId"`
	ClientSecret    string `json:"clientSecret"`
	JwtSecret       string `json:"jwtSecret"`
	JwtExpiry       int    `json:"jwtExpiry"` // in seconds
	AdminID         string `json:"adminId"`
	IPApiKey        string `json:"ipApiKey"`
	MaxHostsPerUser int    `json:"maxHostsPerUser"`
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
