package tomapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Cache management response types

type InvalidateDeviceResponse struct {
	Device       string `json:"device"`
	DeletedCount int    `json:"deleted_count"`
	Message      string `json:"message"`
}

type ClearAllCacheResponse struct {
	DeletedCount int    `json:"deleted_count"`
	Message      string `json:"message"`
}

type ListCacheKeysResponse struct {
	DeviceFilter *string  `json:"device_filter"`
	Count        int      `json:"count"`
	Keys         []string `json:"keys"`
}

type CacheStatsResponse struct {
	Enabled          bool           `json:"enabled"`
	TotalEntries     int            `json:"total_entries"`
	DevicesCached    int            `json:"devices_cached"`
	EntriesPerDevice map[string]int `json:"entries_per_device"`
	DefaultTTL       int            `json:"default_ttl"`
	MaxTTL           int            `json:"max_ttl"`
	KeyPrefix        string         `json:"key_prefix"`
}

// InvalidateDeviceCache invalidates all cache entries for a specific device
func (c *Client) InvalidateDeviceCache(deviceName string) (*InvalidateDeviceResponse, error) {
	apiURL := fmt.Sprintf("%s/api/cache/%s", c.BaseURL, deviceName)

	resp, err := c.makeRequest("DELETE", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(body))
	}

	var result InvalidateDeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ClearAllCache clears all cache entries across all devices
func (c *Client) ClearAllCache() (*ClearAllCacheResponse, error) {
	apiURL := fmt.Sprintf("%s/api/cache", c.BaseURL)

	resp, err := c.makeRequest("DELETE", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(body))
	}

	var result ClearAllCacheResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListCacheKeys lists all cache keys, optionally filtered by device
func (c *Client) ListCacheKeys(deviceName string) (*ListCacheKeysResponse, error) {
	apiURL := fmt.Sprintf("%s/api/cache", c.BaseURL)

	// Add query param if device filter provided
	if deviceName != "" {
		params := url.Values{}
		params.Add("device_name", deviceName)
		apiURL = apiURL + "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(body))
	}

	var result ListCacheKeysResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetCacheStats gets overall cache statistics and configuration
func (c *Client) GetCacheStats() (*CacheStatsResponse, error) {
	apiURL := fmt.Sprintf("%s/api/cache/stats", c.BaseURL)

	resp, err := c.makeRequest("GET", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(body))
	}

	var result CacheStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
