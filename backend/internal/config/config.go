package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	// Level is the minimum log level to display
	// One of: debug, info, warn, error
	Level string `mapstructure:"level"`
	// Format is the log output format
	// One of: json, text
	Format string `mapstructure:"format"`
	// Output is the log output destination
	// One of: stdout, stderr, file
	Output string `mapstructure:"output"`
	// FilePath is the path to the log file (used when Output is "file")
	FilePath string `mapstructure:"file_path"`
	// AddSource adds source file and line number to log output
	AddSource bool `mapstructure:"add_source"`
}

// Config holds all configuration for the application
type Config struct {
	Logger LoggerConfig `mapstructure:"logger"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		Logger: LoggerConfig{
			Level:     "info",
			Format:    "json",
			Output:    "stdout",
			AddSource: true,
		},
	}
}

// Load reads configuration from file and environment variables.
// Environment variables take precedence over file values.
// Environment variable format: LOGGER_LEVEL, LOGGER_FORMAT, etc.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.output", "stdout")
	v.SetDefault("logger.add_source", true)

	// Configure file reading
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./")
		v.AddConfigPath("./etc")
		v.AddConfigPath("./etc/")
		v.AddConfigPath("$HOME/.tradingeino")
	}

	// Configure automatic environment variable binding
	v.SetEnvPrefix("TRADINGEINO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, continue with defaults and env vars
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
