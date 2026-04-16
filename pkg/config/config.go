package config

import (
	"fmt"
)

type Config struct {
	APIKey          string
	BaseURL         string
	Scope           string
	PollingInterval uint32
}

func (cfg *Config) IsValid() error {
	if cfg.APIKey == "" {
		return fmt.Errorf("APIKey is required")
	}
	if cfg.BaseURL == "" {
		return fmt.Errorf("BaseURL is required")
	}
	if cfg.Scope == "" {
		return fmt.Errorf("Scope is required")
	}
	return nil
}

func (cfg *Config) Normalise() {
	if cfg.PollingInterval < 10 {
		cfg.PollingInterval = 10
	}

	if cfg.PollingInterval > 3600 {
		cfg.PollingInterval = 3600
	}
}