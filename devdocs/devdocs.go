// Package devdocs is the library behind the devdocs command: the HTTP client,
// request shaping, and the typed data models for DevDocs.io.
//
// DevDocs.io exposes a fully public JSON API at https://devdocs.io with no
// authentication required. The Client here is the spine every command shares.
// It sets a real User-Agent, paces requests so a busy session stays polite,
// and retries the transient failures (429 and 5xx) that any public site throws
// under load.
package devdocs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DefaultUserAgent identifies the client to DevDocs.io.
const DefaultUserAgent = "devdocs/dev (+https://github.com/tamnd/devdocs-cli)"

// ErrNotFound is returned when the API returns null for a resource.
var ErrNotFound = errors.New("not found")

// Config holds constructor parameters for the DevDocs client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://devdocs.io",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   30 * time.Second,
	}
}

// Client talks to the DevDocs.io JSON API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	rate       time.Duration
	retries    int
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client configured by cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    cfg.BaseURL,
		userAgent:  cfg.UserAgent,
		rate:       cfg.Rate,
		retries:    cfg.Retries,
	}
}

// get fetches a URL with pacing and retries.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// getJSON fetches and JSON-decodes into v. Returns ErrNotFound when the body is null.
func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "null" {
		return ErrNotFound
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

// ListDocs returns all documentation sets available on DevDocs.io.
func (c *Client) ListDocs(ctx context.Context) ([]DocSet, error) {
	rawURL := c.baseURL + "/docs/index.json"
	var wire []wireDocSet
	if err := c.getJSON(ctx, rawURL, &wire); err != nil {
		return nil, err
	}
	out := make([]DocSet, len(wire))
	for i, w := range wire {
		out[i] = wireToDocSet(w)
	}
	return out, nil
}

// DocEntries returns the entry table of contents for a doc set identified by slug.
// The slug may contain ~ (e.g., python~3.11).
func (c *Client) DocEntries(ctx context.Context, slug string) ([]Entry, error) {
	rawURL := c.baseURL + "/docs/" + slug + "/index.json"
	var idx wireIndex
	if err := c.getJSON(ctx, rawURL, &idx); err != nil {
		return nil, err
	}
	return idx.Entries, nil
}

// DocTypes returns the entry type summary for a doc set identified by slug.
func (c *Client) DocTypes(ctx context.Context, slug string) ([]EntryType, error) {
	rawURL := c.baseURL + "/docs/" + slug + "/index.json"
	var idx wireIndex
	if err := c.getJSON(ctx, rawURL, &idx); err != nil {
		return nil, err
	}
	return idx.Types, nil
}
