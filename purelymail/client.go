package purelymail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

type apiEnvelope struct {
	Type    string          `json:"type"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
	Error   *APIError       `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Do(ctx context.Context, operation string, request any, response any) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal %s request: %w", operation, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v0/"+operation, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create %s request: %w", operation, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Purelymail-Api-Token", c.token)
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read %s response: %w", operation, err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var apiErr APIError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("%s failed (%s): %s", operation, apiErr.Code, apiErr.Message)
		}
		return fmt.Errorf("%s failed with HTTP status %s: %s", operation, res.Status, strings.TrimSpace(string(body)))
	}
	var envelope apiEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode %s response: %w", operation, err)
	}
	if envelope.Type == "error" {
		return fmt.Errorf("%s failed (%s): %s", operation, envelope.Code, envelope.Message)
	}
	if envelope.Error != nil {
		return fmt.Errorf("%s failed (%s): %s", operation, envelope.Error.Code, envelope.Error.Message)
	}
	if response != nil && len(envelope.Result) > 0 {
		if err := json.Unmarshal(envelope.Result, response); err != nil {
			return fmt.Errorf("decode %s result: %w", operation, err)
		}
	}
	return nil
}
