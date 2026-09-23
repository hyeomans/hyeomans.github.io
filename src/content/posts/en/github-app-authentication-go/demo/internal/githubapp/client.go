package githubapp

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

// apiVersion pins the REST API behaviour this workshop was written against.
const apiVersion = "2026-03-10"

// APIError carries the HTTP status of a failed GitHub call. Its message
// deliberately omits request headers so tokens and secrets never leak into
// logs or responses.
type APIError struct {
	Status int
	Method string
	Path   string
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("github %s %s returned %d", e.Method, e.Path, e.Status)
}

// httpClient wraps net/http so every call carries the GitHub headers and base
// URL can be overridden for local development.
type httpClient struct {
	client  *http.Client
	apiBase string
}

func newHTTPClient(apiBase string) *httpClient {
	return &httpClient{
		client:  &http.Client{Timeout: 15 * time.Second},
		apiBase: strings.TrimRight(apiBase, "/"),
	}
}

// json calls the REST API with an optional bearer token and JSON body.
func (h *httpClient) json(ctx context.Context, method, url, token string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", apiVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return h.execute(req)
}

func (h *httpClient) execute(req *http.Request) ([]byte, error) {
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return data, &APIError{
			Status: resp.StatusCode,
			Method: req.Method,
			Path:   req.URL.Path,
			Body:   string(data),
		}
	}
	return data, nil
}
