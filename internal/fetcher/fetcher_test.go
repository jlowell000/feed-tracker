package fetcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jlowell000/feed-tracker/internal/config"
)

func newTestFetcher() *Fetcher {
	return New(config.HTTPConfig{Timeout: 5 * time.Second, UserAgent: "test/1.0"})
}

func TestFetchSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "test/1.0" {
			t.Errorf("User-Agent = %q, want %q", r.Header.Get("User-Agent"), "test/1.0")
		}
		w.Header().Set("ETag", `"abc123"`)
		w.Header().Set("Last-Modified", "Mon, 01 Jan 2024 00:00:00 GMT")
		fmt.Fprint(w, "feed content")
	}))
	defer ts.Close()

	f := newTestFetcher()
	result, err := f.Fetch(ts.URL, "", "")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if result.Status != 200 {
		t.Errorf("Status = %d, want 200", result.Status)
	}
	if string(result.Body) != "feed content" {
		t.Errorf("Body = %q, want %q", string(result.Body), "feed content")
	}
	if result.ETag != `"abc123"` {
		t.Errorf("ETag = %q, want %q", result.ETag, `"abc123"`)
	}
	if result.LastModified != "Mon, 01 Jan 2024 00:00:00 GMT" {
		t.Errorf("LastModified = %q, want %q", result.LastModified, "Mon, 01 Jan 2024 00:00:00 GMT")
	}
}

func TestFetchNotModified(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") != `"abc123"` {
			t.Errorf("If-None-Match = %q, want %q", r.Header.Get("If-None-Match"), `"abc123"`)
		}
		w.WriteHeader(http.StatusNotModified)
	}))
	defer ts.Close()

	f := newTestFetcher()
	result, err := f.Fetch(ts.URL, `"abc123"`, "")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if result.Status != 304 {
		t.Errorf("Status = %d, want 304", result.Status)
	}
	if result.Body != nil {
		t.Errorf("Body should be nil on 304")
	}
}

func TestFetchIfModifiedSince(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-Modified-Since") != "Mon, 01 Jan 2024 00:00:00 GMT" {
			t.Errorf("If-Modified-Since = %q, want %q", r.Header.Get("If-Modified-Since"), "Mon, 01 Jan 2024 00:00:00 GMT")
		}
		w.WriteHeader(http.StatusNotModified)
	}))
	defer ts.Close()

	f := newTestFetcher()
	result, err := f.Fetch(ts.URL, "", "Mon, 01 Jan 2024 00:00:00 GMT")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if result.Status != 304 {
		t.Errorf("Status = %d, want 304", result.Status)
	}
}

func TestFetchServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	f := newTestFetcher()
	_, err := f.Fetch(ts.URL, "", "")
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestFetchNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	f := newTestFetcher()
	_, err := f.Fetch(ts.URL, "", "")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestFetchTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(w, "too late")
	}))
	defer ts.Close()

	f := New(config.HTTPConfig{Timeout: 1 * time.Millisecond, UserAgent: "test/1.0"})
	_, err := f.Fetch(ts.URL, "", "")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestFetchPreservesETagOnEmptyResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "content")
	}))
	defer ts.Close()

	f := newTestFetcher()
	result, err := f.Fetch(ts.URL, `"old-etag"`, "")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	// When server doesn't send ETag, preserve the old one
	if result.ETag != `"old-etag"` {
		t.Errorf("ETag = %q, want preserved old etag %q", result.ETag, `"old-etag"`)
	}
}

func TestFetchPreservesLastModifiedOnEmptyResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "content")
	}))
	defer ts.Close()

	f := newTestFetcher()
	result, err := f.Fetch(ts.URL, "", "old-modified")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if result.LastModified != "old-modified" {
		t.Errorf("LastModified = %q, want preserved old %q", result.LastModified, "old-modified")
	}
}
