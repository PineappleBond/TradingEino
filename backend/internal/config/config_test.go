package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Logger.Level != "info" {
		t.Errorf("expected default level 'info', got '%s'", cfg.Logger.Level)
	}
	if cfg.Logger.Output != "stdout" {
		t.Errorf("expected default output 'stdout', got '%s'", cfg.Logger.Output)
	}
	if cfg.Logger.AddSource != true {
		t.Errorf("expected default add_source 'true', got '%v'", cfg.Logger.AddSource)
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	// When no config file exists, should return defaults
	// Pass empty string to use default search paths
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load should not fail when config file doesn't exist: %v", err)
	}

	if cfg.Logger.Level != "info" {
		t.Errorf("expected default level 'info', got '%s'", cfg.Logger.Level)
	}
}

func TestLoad_FromYamlFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
logger:
  level: debug
  output: stderr
  add_source: false
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Logger.Level != "debug" {
		t.Errorf("expected level 'debug', got '%s'", cfg.Logger.Level)
	}
	if cfg.Logger.Output != "stderr" {
		t.Errorf("expected output 'stderr', got '%s'", cfg.Logger.Output)
	}
	if cfg.Logger.AddSource != false {
		t.Errorf("expected add_source 'false', got '%v'", cfg.Logger.AddSource)
	}
}

func TestLoad_EnvironmentVariableOverride(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
logger:
  level: info
  output: stdout
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	// Set environment variables (viper uses TRADINGEINO_LOGGER_LEVEL format)
	t.Setenv("TRADINGEINO_LOGGER_LEVEL", "error")
	t.Setenv("TRADINGEINO_LOGGER_OUTPUT", "file")
	t.Setenv("TRADINGEINO_LOGGER_FILE_PATH", "/tmp/test.log")

	// Create a new viper instance to test env override
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetEnvPrefix("TRADINGEINO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set default for file_path so env var can override it
	v.SetDefault("logger.file_path", "")

	err = v.ReadInConfig()
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var cfg Config
	err = v.Unmarshal(&cfg)
	if err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	// Environment variables should override file values
	if cfg.Logger.Level != "error" {
		t.Errorf("expected level 'error' from env, got '%s'", cfg.Logger.Level)
	}
	if cfg.Logger.Output != "file" {
		t.Errorf("expected output 'file' from env, got '%s'", cfg.Logger.Output)
	}
	if cfg.Logger.FilePath != "/tmp/test.log" {
		t.Errorf("expected file_path '/tmp/test.log' from env, got '%s'", cfg.Logger.FilePath)
	}
}

func TestLoad_InvalidYaml(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Invalid YAML content
	configContent := `
logger:
  level: "unclosed string
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	_, err = Load(configPath)
	if err == nil {
		t.Error("Load should fail with invalid YAML")
	}
}

func TestLoad_WithFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
logger:
  level: warn
  output: file
  file_path: /var/log/app.log
  add_source: true
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Logger.Level != "warn" {
		t.Errorf("expected level 'warn', got '%s'", cfg.Logger.Level)
	}
	if cfg.Logger.Output != "file" {
		t.Errorf("expected output 'file', got '%s'", cfg.Logger.Output)
	}
	if cfg.Logger.FilePath != "/var/log/app.log" {
		t.Errorf("expected file_path '/var/log/app.log', got '%s'", cfg.Logger.FilePath)
	}
	if cfg.Logger.AddSource != true {
		t.Errorf("expected add_source 'true', got '%v'", cfg.Logger.AddSource)
	}
}
