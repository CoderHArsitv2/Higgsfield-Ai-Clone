// Package httpx is the single HTTP client every provider adapter uses, so
// timeouts, retries and error redaction are consistent rather than
// reimplemented per vendor.
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base    string
	http    *http.Client
	headers map[string]string
}

func New(base string, timeout time.Duration) *Client {
	return &Client{
		base:    strings.TrimSuffix(base, "/"),
		http:    &http.Client{Timeout: timeout},
		headers: map[string]string{},
	}
}

func (c *Client) WithHeader(k, v string) *Client {
	clone := &Client{base: c.base, http: c.http, headers: map[string]string{}}
	for k, v := range c.headers {
		clone.headers[k] = v
	}
	clone.headers[k] = v
	return clone
}

// safePath strips the query string except for its parameter names.
//
// Some providers authenticate by query parameter rather than header -- Gemini
// puts the API key in ?key= -- so logging a raw URL would write live
// credentials into the log.
func safePath(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "(unparseable url)"
	}
	if u.RawQuery == "" {
		return u.Path
	}
	names := make([]string, 0, 4)
	for k := range u.Query() {
		names = append(names, k)
	}
	return u.Path + "?[" + strings.Join(names, ",") + "]"
}

// logCall records every outbound provider request, so "did a call actually go
// out, and what came back" is answerable from the log alone.
func (c *Client) logCall(method, path string, status int, took time.Duration, err error) {
	attrs := []any{
		"method", method,
		"url", c.base + safePath(path),
		"ms", took.Milliseconds(),
	}
	if err != nil {
		slog.Default().Error("provider call failed", append(attrs, "error", err.Error())...)
		return
	}
	if status >= 400 {
		slog.Default().Error("provider call rejected", append(attrs, "status", status)...)
		return
	}
	slog.Default().Info("provider call", append(attrs, "status", status)...)
}

type StatusError struct {
	Status int
	Body   string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("provider returned %d: %s", e.Status, e.Body)
}

func (c *Client) JSON(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	started := time.Now()
	res, err := c.http.Do(req)
	if err != nil {
		c.logCall(method, path, 0, time.Since(started), err)
		return err
	}
	defer res.Body.Close()
	c.logCall(method, path, res.StatusCode, time.Since(started), nil)

	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return &StatusError{Status: res.StatusCode, Body: string(raw)}
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(raw, out)
}

// Bytes fetches a raw body. Some providers hand back the generated file itself
// rather than a URL to it, so those adapters download and re-host.
func (c *Client) Bytes(ctx context.Context, method, path string, body any) ([]byte, string, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, "", err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return nil, "", err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	started := time.Now()
	res, err := c.http.Do(req)
	if err != nil {
		c.logCall(method, path, 0, time.Since(started), err)
		return nil, "", err
	}
	defer res.Body.Close()
	c.logCall(method, path, res.StatusCode, time.Since(started), nil)

	raw, err := io.ReadAll(io.LimitReader(res.Body, 200<<20))
	if err != nil {
		return nil, "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, "", &StatusError{Status: res.StatusCode, Body: string(raw)}
	}
	return raw, res.Header.Get("Content-Type"), nil
}
