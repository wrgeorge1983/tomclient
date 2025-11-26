package tomapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SendDeviceCommand sends a command to a device and returns the result
func (c *Client) SendDeviceCommand(deviceName, command string, wait bool, rawOutput bool, useCache bool, cacheTTL *int, cacheRefresh bool) (string, error) {
	apiURL := fmt.Sprintf("%s/api/device/%s/send_command", c.BaseURL, deviceName)

	reqBody := SendCommandRequest{
		Command:      command,
		Wait:         wait,
		UseCache:     useCache,
		CacheTTL:     cacheTTL,
		CacheRefresh: cacheRefresh,
	}

	resp, err := c.makeJSONRequest("POST", apiURL, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status code: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// If rawOutput is enabled and wait is true, extract the command output from the response
	if rawOutput && wait {
		return extractCommandOutput(body, command)
	}

	return string(body), nil
}

// extractCommandOutput extracts raw command output from a JobResponse
func extractCommandOutput(body []byte, command string) (string, error) {
	// First try to parse as a simple string (direct raw output)
	var result string
	if err := json.Unmarshal(body, &result); err == nil {
		return result, nil
	}

	// Try to parse as a JobResponse with command_data
	var jobResp struct {
		Result interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &jobResp); err != nil {
		return string(body), nil
	}

	// Check if result contains command data
	if resultMap, ok := jobResp.Result.(map[string]interface{}); ok {
		if data, ok := resultMap["data"].(map[string]interface{}); ok {
			if output, ok := data[command].(string); ok {
				return output, nil
			}
		}
	}

	return string(body), nil
}

// SendDeviceCommandWithAuth sends a command to a device with credential override
func (c *Client) SendDeviceCommandWithAuth(deviceName, command, username, password string, wait bool, rawOutput bool, timeout int, useCache bool, cacheTTL *int, cacheRefresh bool) (string, error) {
	apiURL := fmt.Sprintf("%s/api/device/%s/send_command", c.BaseURL, deviceName)

	reqBody := SendCommandRequest{
		Command:      command,
		Wait:         wait,
		Timeout:      timeout,
		UseCache:     useCache,
		CacheTTL:     cacheTTL,
		CacheRefresh: cacheRefresh,
	}

	if username != "" {
		reqBody.Username = &username
	}
	if password != "" {
		reqBody.Password = &password
	}

	resp, err := c.makeJSONRequest("POST", apiURL, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status code: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if rawOutput && wait {
		return extractCommandOutput(body, command)
	}

	return string(body), nil
}

// RecordJWT() sends a request to the /dev/record-jwt endpoint to record JWT token for a device
func (c *Client) RecordJWT() error {
	apiURL := fmt.Sprintf("%s/api/dev/record-jwt", c.BaseURL)
	resp, err := c.makeRequest("POST", apiURL)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	return nil
}
