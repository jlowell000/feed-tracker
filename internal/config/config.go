package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	HTTP     HTTPConfig     `yaml:"http"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type HTTPConfig struct {
	Timeout   time.Duration `yaml:"timeout"`
	UserAgent string        `yaml:"user_agent"`
}

func (c *Config) SetDefaults() {
	if c.Database.Path == "" {
		c.Database.Path = "./data/feeds.db"
	}
	if c.HTTP.Timeout == 0 {
		c.HTTP.Timeout = 30 * time.Second
	}
	if c.HTTP.UserAgent == "" {
		c.HTTP.UserAgent = "feed-tracker/0.1"
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.SetDefaults()
	return &cfg, nil
}
