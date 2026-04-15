package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/twospoon/foff-feature-config-go-sdk/pkg/config"
	"github.com/twospoon/foff-feature-config-go-sdk/pkg/models"
)

const (
	apiKeyHeaderKey = "X-FOFF-API-Key"
)

type Client struct {
	config     *config.Config
	httpClient *http.Client

	mu    sync.RWMutex
	cache *models.AllConfigsForScope

	cancel context.CancelFunc
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {

	err := cfg.IsValid()
	if err != nil {
		return nil, fmt.Errorf("foff: invalid config: %w", err)
	}

	c := &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Initial fetch
	if err := c.fetchConfigs(ctx); err != nil {
		return nil, fmt.Errorf("foff: initial config fetch failed: %w", err)
	}

	// Start background polling if interval > 0
	if cfg.PollingInterval > 0 {
		pollCtx, cancel := context.WithCancel(ctx)
		c.cancel = cancel
		go c.poll(pollCtx)
	}

	return c, nil
}

// Close stops the background polling goroutine.
func (c *Client) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Client) poll(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(c.config.PollingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.fetchConfigs(ctx)
		}
	}
}

func (c *Client) fetchConfigs(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/scopes/%s/configs", c.config.BaseURL, c.config.Scope)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("foff: failed to create request: %w", err)
	}

	req.Header.Set(apiKeyHeaderKey, c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("foff: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("foff: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result models.AllConfigsForScope
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("foff: failed to decode response: %w", err)
	}

	c.mu.Lock()
	c.cache = &result
	c.mu.Unlock()

	return nil
}

// GetFeatureConfig returns the config map for a specific feature, or nil if not found.
func (c *Client) GetFeatureConfig(featureName string, orderedHeirarchy []string) interface{} {

	// check if feature exists at all
	c.mu.RLock()
	_, featureExists := c.cache.Features[featureName]
	c.mu.RUnlock()

	if !featureExists {
		return nil
	}


	c.mu.RLock()
	defer c.mu.RUnlock()

	for i := len(orderedHeirarchy); i > 0; i-- {
		heirarchyString := strings.Join(orderedHeirarchy[:i], "#")
		if featureConfig, exists := c.cache.Features[featureName][heirarchyString]; exists {
			return featureConfig
		}
	}

	return c.cache.Features[featureName]["default"]
}
