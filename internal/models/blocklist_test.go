package models

import (
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"
	"testing"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name     string
		statuses []int
		wantErr  bool
		wantHits int32
	}{
		{"ok", []int{http.StatusOK}, false, 1},
		{"retries transient status", []int{http.StatusBadGateway, http.StatusOK}, false, 2},
		{"gives up after attempts", []int{503, 503, 503, 503}, true, fetchAttempts},
		{"no retry on 404", []int{http.StatusNotFound, http.StatusOK}, true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hits atomic.Int32

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				status := tt.statuses[hits.Add(1)-1]
				w.WriteHeader(status)
				_, _ = w.Write([]byte("||a.com^\n"))
			}))
			defer srv.Close()

			b := Blocklist{Name: "test", Source: srv.URL}
			err := b.fetch(t.Context(), FetchOptions{})

			if hits.Load() != tt.wantHits {
				t.Fatalf("expected %d requests, got %d", tt.wantHits, hits.Load())
			}

			if tt.wantErr {
				if !errors.Is(err, ErrBadStatus) {
					t.Fatalf("expected ErrBadStatus, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("fetch: %v", err)
			}

			if string(b.Data) != "||a.com^\n" {
				t.Fatalf("unexpected body %q", b.Data)
			}
		})
	}
}

func TestFetchFallsBackToCache(t *testing.T) {
	var down atomic.Bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if down.Load() {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		_, _ = w.Write([]byte("||good.com^\n"))
	}))
	defer srv.Close()

	opts := FetchOptions{CacheDir: t.TempDir()}

	if err := (&Blocklist{Name: "a", Source: srv.URL}).fetch(t.Context(), opts); err != nil {
		t.Fatalf("first fetch: %v", err)
	}

	down.Store(true)

	b := Blocklist{Name: "a", Source: srv.URL}
	if err := b.fetch(t.Context(), opts); err != nil {
		t.Fatalf("expected cached copy, got %v", err)
	}

	if string(b.Data) != "||good.com^\n" {
		t.Fatalf("unexpected cached body %q", b.Data)
	}

	// Without a cached copy the failure surfaces.
	if err := (&Blocklist{Name: "a", Source: srv.URL}).fetch(t.Context(), FetchOptions{}); err == nil {
		t.Fatal("expected an error without a cache")
	}
}

func TestFetchLocalAndSizeLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.txt")
	if err := os.WriteFile(path, []byte("||local.com^\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	b := Blocklist{Name: "local", Source: path}
	if err := b.fetch(t.Context(), FetchOptions{MaxBytes: 100}); err != nil {
		t.Fatalf("local fetch: %v", err)
	}

	if string(b.Data) != "||local.com^\n" {
		t.Fatalf("unexpected body %q", b.Data)
	}

	err := (&Blocklist{Name: "local", Source: path}).fetch(t.Context(), FetchOptions{MaxBytes: 5})
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestProcessLoadsPatternSources(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "list.txt")
	excl := filepath.Join(dir, "exclusions.txt")

	for path, body := range map[string]string{list: "||a.com^\n", excl: "! comment\nb.com\n"} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	config := BlocklistConfig{
		Name: "t",
		Blocklists: []Blocklist{{
			Name: "l", Source: list, Type: hostsType,
			Filters: Filters{ExclusionsSources: []string{excl}},
		}},
		Filters: Filters{Exclusions: []string{"c.com"}, InclusionsSources: []string{excl}},
	}

	if err := config.Process(t.Context(), FetchOptions{}); err != nil {
		t.Fatalf("process: %v", err)
	}

	if got := config.Blocklists[0].Exclusions; !slices.Contains(got, "b.com") {
		t.Fatalf("source exclusions not loaded: %q", got)
	}

	if !slices.Contains(config.Exclusions, "c.com") || !slices.Contains(config.Inclusions, "b.com") {
		t.Fatalf("config filters not loaded: %q / %q", config.Exclusions, config.Inclusions)
	}

	config.InclusionsSources = []string{filepath.Join(dir, "missing.txt")}
	if err := config.Process(t.Context(), FetchOptions{}); err == nil {
		t.Fatal("expected an error for a missing pattern source")
	}
}

func TestFiltersDecodeInline(t *testing.T) {
	var config BlocklistConfig

	err := json.Unmarshal([]byte(`{
		"name": "t", "exclusions": ["a"], "inclusions_sources": ["b"],
		"sources": [{"name": "s", "source": "x", "type": "hosts", "exclusions_sources": ["c"]}]
	}`), &config)
	if err != nil {
		t.Fatal(err)
	}

	if len(config.Exclusions) != 1 || len(config.InclusionsSources) != 1 ||
		len(config.Blocklists[0].ExclusionsSources) != 1 {
		t.Fatalf("filters not decoded: %+v", config)
	}
}
