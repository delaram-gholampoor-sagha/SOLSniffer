package configs

import (
	"fmt"
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/utils"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

type RetryConfig struct {
	Attempts  uint          `yaml:"attempts"`
	Delay     time.Duration `yaml:"delay"`
	DelayType string        `yaml:"delay_type"`
}

type DatabaseConfig struct {
	URI   string      `yaml:"uri"`
	Retry RetryConfig `yaml:"retry"`
}

type WebSocketConfig struct {
	Scheme string      `yaml:"scheme"`
	Host   string      `yaml:"host"`
	Path   string      `yaml:"path"`
	Retry  RetryConfig `yaml:"retry"`
}

type ServicesConfig struct {
	Wallets []string `yaml:"wallets"`
	Tokens  []string `yaml:"tokens"`
}

type CoordinatorConfig struct {
	Retry RetryConfig `yaml:"retry"`
}

type AppConfig struct {
	Env             utils.Environment `yaml:"env"`
	Addr            string            `yaml:"addr"`
	ApplicationName string            `yaml:"application_name"`
	Log             LogConfig         `yaml:"log"`
}

type LogConfig struct {
	LogLevel    string `yaml:"log_level"`
	PrettyPrint bool   `yaml:"pretty_print"`
}

type Config struct {
	App         AppConfig         `yaml:"app"`
	Database    DatabaseConfig    `yaml:"database"`
	WebSocket   WebSocketConfig   `yaml:"websocket"`
	Services    ServicesConfig    `yaml:"services"`
	Coordinator CoordinatorConfig `yaml:"coordinator"`
}

// Load reads and parses the YAML configuration file.
func Load(configPath string) (*Config, error) {
	var cfg Config

	// Get the absolute path to the configuration file
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	// Open the configuration file
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Decode the YAML into the Config struct
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	// Validate required fields
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Database.URI == "" {
		return fmt.Errorf("database.uri is required")
	}
	if cfg.WebSocket.Scheme == "" {
		return fmt.Errorf("websocket.scheme is required")
	}
	if cfg.WebSocket.Host == "" {
		return fmt.Errorf("websocket.host is required")
	}
	if cfg.WebSocket.Path == "" {
		return fmt.Errorf("websocket.path is required")
	}
	if len(cfg.Services.Wallets) == 0 {
		return fmt.Errorf("services.wallets must have at least one entry")
	}
	if len(cfg.Services.Tokens) == 0 {
		return fmt.Errorf("services.tokens must have at least one entry")
	}
	return nil
}
