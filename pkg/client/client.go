package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/twospoon/foff-feature-config-go-sdk/pkg/config"
	"github.com/twospoon/foff-feature-config-go-sdk/pkg/models"
)

const (
	apiKeyHeaderKey = "X-FOFF-API-Key"
	emailHeaderKey  = "X-FOFF-Email"
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
		pollCtx, cancel := context.WithCancel(context.Background())
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
	req.Header.Set(emailHeaderKey, c.config.Email)

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

// GetAllConfigs returns the cached config response.
func (c *Client) GetAllConfigs() *models.AllConfigsForScope {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache
}

// GetFeatureConfig returns the config map for a specific feature, or nil if not found.
func (c *Client) GetFeatureConfig(featureName string) map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache == nil {
		return nil
	}
	return c.cache.Features[featureName]
}

// GetFeatureConfigValue returns a specific key's value for a feature, or nil if not found.
func (c *Client) GetFeatureConfigValue(featureName, key string) interface{} {
	featureCfg := c.GetFeatureConfig(featureName)
	if featureCfg == nil {
		return nil
	}
	return featureCfg[key]
}

// GetOrderedHierarchy returns the ordered hierarchy from the cached response.
func (c *Client) GetOrderedHierarchy() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache == nil {
		return nil
	}
	return c.cache.OrderedHeirarchy
}
