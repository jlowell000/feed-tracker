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
	Prune    PruneConfig    `yaml:"prune"`
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

type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	v, err := parseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

func (d Duration) MarshalYAML() (any, error) {
	return time.Duration(d).String(), nil
}

func parseDuration(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

type PruneConfig struct {
	MaxAge Duration `yaml:"max_age"`
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
