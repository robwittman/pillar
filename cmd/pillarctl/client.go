package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *APIClient) Get(path string) ([]byte, int, error) {
	return c.do("GET", path, nil)
}

func (c *APIClient) Post(path string, body any) ([]byte, int, error) {
	return c.do("POST", path, body)
}

func (c *APIClient) Put(path string, body any) ([]byte, int, error) {
	return c.do("PUT", path, body)
}

func (c *APIClient) Delete(path string) (int, error) {
	_, code, err := c.do("DELETE", path, nil)
	return code, err
}

func (c *APIClient) do(method, path string, body any) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, resp.StatusCode, fmt.Errorf("server error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, resp.StatusCode, fmt.Errorf("server error (%d)", resp.StatusCode)
	}

	return respBody, resp.StatusCode, nil
}
