package config

import (
	"fmt"
)

type Config struct {
	APIKey          string
	Email           string
	BaseURL         string
	Scope           string
	PollingInterval uint32
}

func (cfg *Config) IsValid() error {
	if cfg.APIKey == "" {
		return fmt.Errorf("APIKey is required")
	}
	if cfg.Email == "" {
		return fmt.Errorf("Email is required")
	}
	if cfg.BaseURL == "" {
		return fmt.Errorf("BaseURL is required")
	}
	if cfg.Scope == "" {
		return fmt.Errorf("Scope is required")
	}
	return nil
}
