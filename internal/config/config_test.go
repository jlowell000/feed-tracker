package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	yaml := `database:
  path: /tmp/test.db
http:
  timeout: 10s
  user_agent: "test/1.0"
`
	f, err := os.CreateTemp(t.TempDir(), "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f.WriteString(yaml)

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Database.Path != "/tmp/test.db" {
		t.Errorf("Database.Path = %q, want %q", cfg.Database.Path, "/tmp/test.db")
	}
	if cfg.HTTP.Timeout != 10*time.Second {
		t.Errorf("HTTP.Timeout = %v, want %v", cfg.HTTP.Timeout, 10*time.Second)
	}
	if cfg.HTTP.UserAgent != "test/1.0" {
		t.Errorf("HTTP.UserAgent = %q, want %q", cfg.HTTP.UserAgent, "test/1.0")
	}
}

func TestSetDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefaults()
	if cfg.Database.Path == "" {
		t.Error("Database.Path should have default")
	}
	if cfg.HTTP.Timeout == 0 {
		t.Error("HTTP.Timeout should have default")
	}
	if cfg.HTTP.UserAgent == "" {
		t.Error("HTTP.UserAgent should have default")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
