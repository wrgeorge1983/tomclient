package tomapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// ExportInventory exports inventory in DeviceConfig format
func (c *Client) ExportInventory(filter string) (map[string]DeviceConfig, error) {
	apiURL := c.BaseURL + "/api/inventory/export"
	if filter != "" {
		apiURL += "?filter_name=" + url.QueryEscape(filter)
	}

	resp, err := c.makeRequest("GET", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var inventory map[string]DeviceConfig
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return inventory, nil
}

// ExportRawInventory exports raw inventory nodes
func (c *Client) ExportRawInventory(filter string) ([]RawInventoryNode, error) {
	apiURL := c.BaseURL + "/api/inventory/export/raw"
	if filter != "" {
		apiURL += "?filter_name=" + url.QueryEscape(filter)
	}

	resp, err := c.makeRequest("GET", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var inventory []RawInventoryNode
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return inventory, nil
}

// GetDeviceConfig gets configuration for a specific device
func (c *Client) GetDeviceConfig(deviceName string) (*DeviceConfig, error) {
	apiURL := fmt.Sprintf("%s/api/inventory/%s", c.BaseURL, deviceName)

	resp, err := c.makeRequest("GET", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var config DeviceConfig
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &config, nil
}

// ExportInventoryWithFilters exports inventory with inline field filters
func (c *Client) ExportInventoryWithFilters(filters map[string]string) (map[string]DeviceConfig, error) {
	apiURL := c.BaseURL + "/api/inventory/export"

	if len(filters) > 0 {
		params := url.Values{}
		for field, pattern := range filters {
			params.Add(field, pattern)
		}
		apiURL += "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var inventory map[string]DeviceConfig
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return inventory, nil
}

// ListFilters gets available inventory filters
func (c *Client) ListFilters() (map[string]string, error) {
	apiURL := c.BaseURL + "/api/inventory/filters"

	resp, err := c.makeRequest("GET", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var filters map[string]string
	err = json.Unmarshal(body, &filters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return filters, nil
}
