package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jlowell000/feed-tracker/internal/config"
)

type Result struct {
	Body         []byte
	ETag         string
	LastModified string
	Status       int
}

type Fetcher struct {
	client *http.Client
	ua     string
}

func New(cfg config.HTTPConfig) *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: cfg.Timeout},
		ua:     cfg.UserAgent,
	}
}

func (f *Fetcher) Fetch(url, etag, lastModified string) (*Result, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", f.ua)
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return &Result{Status: resp.StatusCode}, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	result := &Result{
		Body:         body,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		Status:       resp.StatusCode,
	}

	if result.ETag == "" {
		result.ETag = etag
	}
	if result.LastModified == "" {
		result.LastModified = lastModified
	}

	return result, nil
}

func (f *Fetcher) FetchWithTimeout(url string, timeout time.Duration) (*Result, error) {
	f.client.Timeout = timeout
	return f.Fetch(url, "", "")
}
