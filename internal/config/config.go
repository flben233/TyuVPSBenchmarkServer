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
	AccessTokenExp  int    `json:"accessTokenExp"`  // in seconds
	RefreshTokenExp int    `json:"refreshTokenExp"` // in seconds
	AdminID         int64  `json:"adminId"`
	FrontendURL     string `json:"frontendUrl"`
	GithubHttpProxy string `json:"githubHttpsProxy"`
	ExporterURL     string `json:"exporterUrl"`
	RedisHost       string `json:"redisHost"`
	RedisPasswd     string `json:"redisPasswd"`
	AppriseURL      string `json:"appriseUrl"`
	KafkaURL        string `json:"kafkaUrl"`
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
