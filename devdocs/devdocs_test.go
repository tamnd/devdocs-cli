package devdocs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/devdocs-cli/devdocs"
)

func TestListDocs(t *testing.T) {
	const fixture = `[
		{
			"name": "Python",
			"slug": "python~3.11",
			"type": "sphinx",
			"links": {"home": "https://python.org", "code": "https://github.com/python/cpython"},
			"version": "3.11",
			"release": "3.11.9",
			"mtime": 1715000000,
			"db_size": 8765432
		},
		{
			"name": "Go",
			"slug": "go",
			"type": "godoc",
			"links": {"home": "https://golang.org", "code": "https://github.com/golang/go"},
			"version": "",
			"release": "1.22.3",
			"mtime": 1715100000,
			"db_size": 4321098
		}
	]`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/docs.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fixture))
	}))
	defer ts.Close()

	cfg := devdocs.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	c := devdocs.NewClient(cfg)

	docs, err := c.ListDocs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("got %d docs, want 2", len(docs))
	}

	py := docs[0]
	if py.Name != "Python" {
		t.Errorf("Name = %q, want %q", py.Name, "Python")
	}
	if py.Slug != "python~3.11" {
		t.Errorf("Slug = %q, want %q", py.Slug, "python~3.11")
	}
	if py.Type != "sphinx" {
		t.Errorf("Type = %q, want %q", py.Type, "sphinx")
	}
	if py.Home != "https://python.org" {
		t.Errorf("Home = %q, want %q", py.Home, "https://python.org")
	}
	if py.Code != "https://github.com/python/cpython" {
		t.Errorf("Code = %q, want %q", py.Code, "https://github.com/python/cpython")
	}
}

func TestDocEntries(t *testing.T) {
	const fixture = `{
		"entries": [
			{"name": "os.path.join", "path": "library/os.path#join", "type": "function"},
			{"name": "str", "path": "library/stdtypes#str", "type": "class"},
			{"name": "open", "path": "library/functions#open", "type": "function"}
		],
		"types": [
			{"name": "function", "count": 432},
			{"name": "class", "count": 89}
		]
	}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/docs/python~3.11/index.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fixture))
	}))
	defer ts.Close()

	cfg := devdocs.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	c := devdocs.NewClient(cfg)

	entries, err := c.DocEntries(context.Background(), "python~3.11")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	first := entries[0]
	if first.Name != "os.path.join" {
		t.Errorf("Name = %q, want %q", first.Name, "os.path.join")
	}
	if first.Path != "library/os.path#join" {
		t.Errorf("Path = %q, want %q", first.Path, "library/os.path#join")
	}
	if first.Type != "function" {
		t.Errorf("Type = %q, want %q", first.Type, "function")
	}
}

func TestDocTypes(t *testing.T) {
	const fixture = `{
		"entries": [
			{"name": "os.path.join", "path": "library/os.path#join", "type": "function"}
		],
		"types": [
			{"name": "function", "count": 432},
			{"name": "class", "count": 89},
			{"name": "method", "count": 1204}
		]
	}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/docs/python~3.11/index.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fixture))
	}))
	defer ts.Close()

	cfg := devdocs.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	c := devdocs.NewClient(cfg)

	types, err := c.DocTypes(context.Background(), "python~3.11")
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 3 {
		t.Fatalf("got %d types, want 3", len(types))
	}

	if types[0].Name != "function" {
		t.Errorf("Name = %q, want %q", types[0].Name, "function")
	}
	if types[0].Count != 432 {
		t.Errorf("Count = %d, want 432", types[0].Count)
	}
}

func TestClientRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	cfg := devdocs.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := devdocs.NewClient(cfg)

	start := time.Now()
	docs, err := c.ListDocs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Errorf("got %d docs after recovery, want 0", len(docs))
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestClientSendsUserAgent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	cfg := devdocs.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	c := devdocs.NewClient(cfg)

	if _, err := c.ListDocs(context.Background()); err != nil {
		t.Fatal(err)
	}
}
