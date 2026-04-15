package basic
package main

import (





































}	fmt.Printf("enabled = %v\n", val)	val := c.GetFeatureConfigValue("my-feature", "enabled")	// Get a specific value from a feature config	fmt.Printf("Feature config: %v\n", featureCfg)	featureCfg := c.GetFeatureConfig("my-feature")	// Get a specific feature config	fmt.Printf("Hierarchy: %v\n", allConfigs.OrderedHeirarchy)	allConfigs := c.GetAllConfigs()	// Get all configs	defer c.Close()	}		log.Fatalf("failed to create client: %v", err)	if err != nil {	c, err := client.NewClient(ctx, cfg)	}		PollingInterval: 30, // seconds		Scope:           "production",		BaseURL:         "https://api.foff.dev",		Email:           "you@example.com",		APIKey:          "your-api-key",	cfg := &config.Config{	ctx := context.Background()func main() {)	"github.com/twospoon/foff-feature-config-go-sdk/pkg/config"	"github.com/twospoon/foff-feature-config-go-sdk/pkg/client"	"log"	"fmt"	"context"