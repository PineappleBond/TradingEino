package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	// Level is the minimum log level to display
	// One of: debug, info, warn, error
	Level string `mapstructure:"level"`
	// Output is the log output destination
	// One of: stdout, stderr, file
	Output string `mapstructure:"output"`
	// FilePath is the path to the log file (used when Output is "file")
	FilePath string `mapstructure:"file_path"`
	// AddSource adds source file and line number to log output
	AddSource bool `mapstructure:"add_source"`
}

// DBConfig holds db configuration
type DBConfig struct {
	Type   string `mapstructure:"type"` // support sqlite only
	DBPath string `mapstructure:"db_path"`
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	// Enabled enables the scheduler
	Enabled bool `mapstructure:"enabled"`
	// MaxConcurrency is the maximum number of concurrent task executions
	MaxConcurrency int `mapstructure:"max_concurrency"`
	// CheckInterval is the interval in seconds to check for due tasks
	CheckInterval int `mapstructure:"check_interval"`
	// DefaultTimeout is the default timeout in seconds for task execution
	DefaultTimeout int `mapstructure:"default_timeout"`
}

// ChatModelConfig holds chat model configuration
type ChatModelConfig struct {
	// APIKey is the API key for the chat model provider
	APIKey string `mapstructure:"api_key"`
	// BaseURL is the base URL for the chat model API
	BaseURL string `mapstructure:"base_url"`
	// Model is the model name to use
	Model string `mapstructure:"model"`
}

// OKXConfig holds OKX exchange configuration
type OKXConfig struct {
	// ApiKey is the OKX API key
	ApiKey string `mapstructure:"api_key"`
	// SecretKey is the OKX API secret key
	SecretKey string `mapstructure:"secret_key"`
	// Passphrase is the OKX API passphrase
	Passphrase string `mapstructure:"passphrase"`
	// Sandbox enables sandbox mode (uses Demo server)
	Sandbox bool `mapstructure:"sandbox"`
}

// Config holds all configuration for the application
type Config struct {
	Logger    LoggerConfig    `mapstructure:"logger"`
	DB        DBConfig        `mapstructure:"db"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	ChatModel ChatModelConfig `mapstructure:"chat_model"`
	OKX       OKXConfig       `mapstructure:"okx"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		Logger: LoggerConfig{
			Level:     "info",
			Output:    "stdout",
			AddSource: true,
		},
		DB: DBConfig{
			Type:   "sqlite",
			DBPath: "./data/TradingEino.db",
		},
		Scheduler: SchedulerConfig{
			Enabled:        true,
			MaxConcurrency: 5,
			CheckInterval:  10,
			DefaultTimeout: 300,
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
	v.SetDefault("db.type", "sqlite")
	v.SetDefault("db.db_path", "./data/TradingEino.db")
	v.SetDefault("scheduler.enabled", true)
	v.SetDefault("scheduler.max_concurrency", 5)
	v.SetDefault("scheduler.check_interval", 10)
	v.SetDefault("scheduler.default_timeout", 300)

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

func (c LoggerConfig) DBLogPath() string {
	filePath := c.FilePath
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	dbLogFile := filepath.Join(dir, "gorm.log"+ext)
	return dbLogFile
}
