package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type InventoryCache struct {
	Devices   []string  `json:"devices"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func GetCachePath(configDir string) string {
	if configDir == "" {
		configDir = GetConfigDir()
	}
	return filepath.Join(configDir, "inventory_cache.json")
}

func GetCacheTTL() time.Duration {
	if ttl := os.Getenv("TOM_CACHE_TTL"); ttl != "" {
		if d, err := time.ParseDuration(ttl); err == nil {
			return d
		}
	}
	return 1 * time.Hour
}

func LoadInventoryCache(configDir string) (*InventoryCache, error) {
	cachePath := GetCachePath(configDir)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache InventoryCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	if time.Now().After(cache.ExpiresAt) {
		return nil, nil
	}

	return &cache, nil
}

func SaveInventoryCache(configDir string, devices []string, ttl time.Duration) error {
	if configDir == "" {
		configDir = GetConfigDir()
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cache := InventoryCache{
		Devices:   devices,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	cachePath := GetCachePath(configDir)
	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

func ClearInventoryCache(configDir string) error {
	cachePath := GetCachePath(configDir)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}
	return nil
}
