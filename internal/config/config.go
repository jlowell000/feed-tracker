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
	TUI      TUIConfig      `yaml:"tui"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type HTTPConfig struct {
	Timeout          time.Duration `yaml:"timeout"`
	UserAgent        string        `yaml:"user_agent"`
	FetchConcurrency int           `yaml:"fetch_concurrency"`
	FetchCooldown    time.Duration `yaml:"fetch_cooldown"`
}

type TUIConfig struct {
	EntryLimit   int           `yaml:"entry_limit"`
	AutoRefresh  time.Duration `yaml:"auto_refresh"`
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
	if c.HTTP.FetchConcurrency <= 0 {
		c.HTTP.FetchConcurrency = 3
	}
	if c.TUI.EntryLimit <= 0 {
		c.TUI.EntryLimit = 100
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
