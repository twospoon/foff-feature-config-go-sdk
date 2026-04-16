# FOFF Golang SDK

This document covers the details of how to integrate with the FOFF Feature Config Service using the Go SDK.

## Installation

```bash
go get github.com/twospoon/foff-feature-config-go-sdk
```

Requires **Go 1.22.10** or later.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/twospoon/foff-feature-config-go-sdk/pkg/client"
	"github.com/twospoon/foff-feature-config-go-sdk/pkg/config"
)

func main() {
	ctx := context.Background()

	cfg := &config.Config{
		APIKey:          "your-api-key",
		BaseURL:         "https://foff.twospoon.ai/live",
		Scope:           "name-of-your-scope",
		PollingInterval: 30, // seconds
	}

	c, err := client.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer c.Close()

	// Retrieve a feature config with hierarchy-based resolution
    // provide the values for the heirarchy for the given user
	value := c.GetFeatureConfig("my-feature", []string{"org-1", "team-a", "user-123"})
	fmt.Println(value)
}
```

## Configuration

The SDK is configured via the `config.Config` struct:

| Field             | Type     | Required | Description                                                                 |
|-------------------|----------|----------|-----------------------------------------------------------------------------|
| `APIKey`          | `string` | Yes      | Your FOFF API key. Sent as the `X-FOFF-API-Key` header on every request.   |
| `BaseURL`         | `string` | Yes      | Base URL of the FOFF API (e.g. `https://api.foff.dev`).                    |
| `Scope`           | `string` | Yes      | The scope to fetch configs for (e.g. `production`, `staging`).             |
| `PollingInterval` | `uint32` | No       | How often (in seconds) to poll for config updates. See [Polling](#polling). |

### Validation

Calling `NewClient` automatically validates the config. It returns an error if any required field is empty.

You can also validate independently:

```go
if err := cfg.IsValid(); err != nil {
    // handle invalid config
}
```

### Normalisation

The polling interval is clamped to safe bounds automatically:

- Values below **10** seconds are raised to **10**.
- Values above **3600** seconds (1 hour) are lowered to **3600**.

## Creating a Client

```go
c, err := client.NewClient(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer c.Close()
```

`NewClient` performs the following steps:

1. Validates the config.
2. Normalises the polling interval.
3. Makes an **initial fetch** of all configs for the given scope from `GET {BaseURL}/api/v1/scopes/{scope}/configs`.
4. If `PollingInterval > 0`, starts a background goroutine that re-fetches configs on the configured interval.

The initial fetch is **blocking** — if the API is unreachable or returns an error, `NewClient` returns that error so you can handle it at startup.

Always call `c.Close()` when you are done to stop the background polling goroutine.

## Fetching Feature Configs

### `GetFeatureConfig(featureName string, orderedHierarchy []string) interface{}`

Returns the config value for a feature, resolved against a hierarchy.

```go
value := c.GetFeatureConfig("dark-mode", []string{"org-1", "team-a", "user-123"})
```

**Hierarchy resolution** works from most-specific to least-specific:

1. The SDK joins the hierarchy elements with `#` and looks up from the most specific combination first.
2. For the example above, it checks keys in this order:
   - `org-1#team-a#user-123` (full match)
   - `org-1#team-a` (two levels)
   - `org-1` (one level)
3. If none match, it falls back to the `"default"` key.
4. If the feature does not exist at all, it returns `nil`.

This lets you define config overrides at any level of your hierarchy (organisation → team → user, environment → region → service, etc.) and the SDK resolves the most specific value automatically.

## Polling

When `PollingInterval` is set, the SDK polls the API in the background and updates its in-memory cache. All reads via `GetFeatureConfig` are served from the cache and are **thread-safe** (protected by a `sync.RWMutex`). By default polling is set to 10 seconds.

- Polling errors are silently ignored — the SDK continues serving the last successfully fetched config.
- The polling goroutine respects the context passed to `NewClient`. Cancelling that context (or calling `Close()`) stops polling.

### Recommended Intervals

| Use Case             | Interval  |
|----------------------|-----------|
| Near-real-time       | 10s       |
| Standard             | 30–60s    |
| Low-traffic / batch  | 300–600s  |

## Thread Safety

All public methods on `Client` are safe for concurrent use. The internal cache is protected by a read-write mutex, so multiple goroutines can call `GetFeatureConfig` simultaneously without contention.



