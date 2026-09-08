package dojoclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a thin REST client for one DefectDojo instance. It carries no
// credential of its own; every call takes the Authorization header value
// to use, since that varies per caller (HTTP mode) or is fixed for the
// life of the process (stdio mode) — see internal/mcptools.authHeader.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

// APIError is returned when DefectDojo responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("defectdojo API returned %d: %s", e.StatusCode, e.Body)
}

// Get performs a GET request against path (e.g. "/api/v2/findings/") with
// the given query parameters and decodes the JSON response into out.
func (c *Client) Get(ctx context.Context, authHeader, path string, query url.Values, out any) error {
	return c.Do(ctx, http.MethodGet, authHeader, path, query, nil, out)
}

// Do performs an arbitrary REST call. Used directly by the generic
// dispatch tool, where the method and body vary per operation and aren't
// known until request time.
func (c *Client) Do(ctx context.Context, method, authHeader, path string, query url.Values, body json.RawMessage, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reqBody io.Reader
	if len(body) > 0 {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling defectdojo: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading defectdojo response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decoding defectdojo response: %w", err)
	}
	return nil
}
