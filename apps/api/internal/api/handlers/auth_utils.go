package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// PostJSON performs an HTTP POST with a JSON body.
func PostJSON(ctx context.Context, url string, body any, resp any) error {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(reqBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close() //nolint:errcheck
	rb, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(rb))
	}
	return json.Unmarshal(rb, resp)
}