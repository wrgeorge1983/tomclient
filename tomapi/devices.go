package tomapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// SendDeviceCommand sends a command to a device and returns the result
func (c *Client) SendDeviceCommand(deviceName, command string, wait bool, rawOutput bool) (string, error) {
	apiURL := fmt.Sprintf("%s/api/device/%s/send_command", c.BaseURL, deviceName)
	
	params := url.Values{}
	params.Add("command", command)
	if wait {
		params.Add("wait", "true")
	}
	if rawOutput {
		params.Add("rawOutput", "true")
	}
	
	fullURL := apiURL + "?" + params.Encode()
	
	resp, err := c.makeRequest("GET", fullURL)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	// If rawOutput is enabled and wait is true, response should be a JSON string
	if rawOutput && wait {
		var result string
		err = json.Unmarshal(body, &result)
		if err != nil {
			// If JSON parsing fails, return raw body
			return string(body), nil
		}
		return result, nil
	}
	
	return string(body), nil
}

// SendDeviceCommandWithAuth sends a command to a device with credential override
func (c *Client) SendDeviceCommandWithAuth(deviceName, command, username, password string, wait bool, rawOutput bool, timeout int) (string, error) {
	apiURL := fmt.Sprintf("%s/api/device/%s/send_command", c.BaseURL, deviceName)
	
	params := url.Values{}
	params.Add("command", command)
	if wait {
		params.Add("wait", "true")
	}
	if rawOutput {
		params.Add("rawOutput", "true")
	}
	if username != "" {
		params.Add("username", username)
	}
	if password != "" {
		params.Add("password", password)
	}
	if timeout > 0 {
		params.Add("timeout", fmt.Sprintf("%d", timeout))
	}
	
	fullURL := apiURL + "?" + params.Encode()
	
	resp, err := c.makeRequest("GET", fullURL)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	if rawOutput && wait {
		var result string
		err = json.Unmarshal(body, &result)
		if err != nil {
			return string(body), nil
		}
		return result, nil
	}
	
	return string(body), nil
}