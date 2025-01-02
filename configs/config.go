package configs

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	MongoURI        string   `yaml:"mongo_uri"`
	SolanaRPC       string   `yaml:"solana_rpc"`
	WebSocketScheme string   `yaml:"websocket_scheme"`
	WebSocketHost   string   `yaml:"websocket_host"`
	WebSocketPath   string   `yaml:"websocket_path"`
	Wallets         []string `yaml:"monitored_wallets"`
	Tokens          []string `yaml:"monitored_tokens"`
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
	if cfg.MongoURI == "" {
		return fmt.Errorf("mongo_uri is required")
	}
	if cfg.SolanaRPC == "" {
		return fmt.Errorf("solana_rpc is required")
	}
	if cfg.WebSocketScheme == "" {
		return fmt.Errorf("websocket_scheme is required")
	}
	if cfg.WebSocketHost == "" {
		return fmt.Errorf("websocket_host is required")
	}
	if cfg.WebSocketPath == "" {
		return fmt.Errorf("websocket_path is required")
	}
	if len(cfg.Wallets) == 0 {
		return fmt.Errorf("monitored_wallets must have at least one entry")
	}
	if len(cfg.Tokens) == 0 {
		return fmt.Errorf("monitored_tokens must have at least one entry")
	}
	return nil
}
