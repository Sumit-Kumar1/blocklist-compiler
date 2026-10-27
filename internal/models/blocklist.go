package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type BlocklistType string

const (
	adblockType BlocklistType = "adblock"
	hostsType   BlocklistType = "hosts"
)

const (
	fetchAttempts = 3
	// fetchTimeout bounds one attempt, body included, so it must allow a large
	// list to download over a slow link.
	fetchTimeout = 2 * time.Minute
	fetchBackoff = 300 * time.Millisecond

	cacheDirPerm  = 0o750
	cacheFilePerm = 0o600
)

var retryStatuses = []int{
	http.StatusTooEarly,
	http.StatusTooManyRequests,
	http.StatusInternalServerError,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
}

// FetchOptions controls how sources are downloaded.
type FetchOptions struct {
	// CacheDir keeps each remote source's last good copy. Empty disables it.
	CacheDir string
	// MaxBytes rejects a source larger than this. Zero means no limit.
	MaxBytes int64
}

type Blocklist struct {
	Name            string        `json:"name"`
	Source          string        `json:"source"`
	Type            BlocklistType `json:"type"`
	Transformations []string      `json:"transformations"`
	Filters
	// Data is the downloaded body, set by fetch.
	Data []byte `json:"-"`
}

func (b Blocklist) validate() error {
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("%w: name in blocklist", ErrMissing)
	}

	if strings.TrimSpace(b.Source) == "" {
		return BlockListError{BlocklistName: b.Name, Err: fmt.Errorf("%w: source", ErrMissing)}
	}

	if b.Type != hostsType && b.Type != adblockType {
		return BlockListError{BlocklistName: b.Name, Err: fmt.Errorf("%w: type %q", ErrInvalid, b.Type)}
	}

	return nil
}

// fetch loads the source into b.Data. A local path is read directly. A URL is
// downloaded with retries, and when every attempt fails the last good copy
// from the cache is used instead, so one unreachable source does not fail the
// whole compile.
func (b *Blocklist) fetch(ctx context.Context, opts FetchOptions) error {
	logger := slog.With(slog.String("blocklist", b.Name))

	if !isRemote(b.Source) {
		data, err := readLocal(b.Source, opts.MaxBytes)
		if err != nil {
			return err
		}

		b.Data = data

		logger.LogAttrs(ctx, slog.LevelInfo, "read local blocklist", slog.Int("bytes", len(data)))

		return nil
	}

	err := b.download(ctx, logger, opts.MaxBytes)
	cache := cachePath(opts.CacheDir, b.Source)

	if err == nil {
		if cache != "" {
			writeCache(ctx, logger, cache, b.Data)
		}

		return nil
	}

	if cache == "" || ctx.Err() != nil {
		return err
	}

	data, cacheErr := os.ReadFile(filepath.Clean(cache))
	if cacheErr != nil {
		return err
	}

	b.Data = data

	logger.LogAttrs(ctx, slog.LevelWarn, "fetch failed, using last good copy",
		slog.String("error", err.Error()), slog.Int("bytes", len(data)))

	return nil
}

// download fetches the URL source, retrying transient failures.
func (b *Blocklist) download(ctx context.Context, logger *slog.Logger, maxBytes int64) error {
	logger.LogAttrs(ctx, slog.LevelInfo, "started fetching blocklist")

	for attempt := 1; ; attempt++ {
		retry, err := b.fetchOnce(ctx, maxBytes)
		if err == nil {
			logger.LogAttrs(ctx, slog.LevelInfo, "fetched blocklist", slog.Int("bytes", len(b.Data)))

			return nil
		}

		if !retry || attempt == fetchAttempts {
			logger.LogAttrs(ctx, slog.LevelError, "fetching blocklist failed", slog.String("error", err.Error()))

			return err
		}

		logger.LogAttrs(ctx, slog.LevelWarn, "retrying fetching blocklist", slog.String("error", err.Error()))

		select {
		case <-ctx.Done():
			return fmt.Errorf("fetch cancelled: %w", ctx.Err())
		case <-time.After(time.Duration(attempt) * fetchBackoff):
		}
	}
}

// fetchOnce performs a single download attempt and reports whether the failure
// it returns is worth retrying.
func (b *Blocklist) fetchOnce(ctx context.Context, maxBytes int64) (bool, error) {
	c, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(c, http.MethodGet, b.Source, http.NoBody)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// This attempt's own timeout is worth another try; a cancelled parent is not.
		return ctx.Err() == nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return slices.Contains(retryStatuses, resp.StatusCode), fmt.Errorf("%w %s", ErrBadStatus, resp.Status)
	}

	data, err := readLimited(resp.Body, maxBytes)
	if err != nil {
		// A body cut off mid-stream is as transient as a failed connect; an
		// oversized one is not.
		return ctx.Err() == nil && !errors.Is(err, ErrTooLarge), err
	}

	b.Data = data

	return false, nil
}

func isRemote(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

func readLocal(path string, maxBytes int64) ([]byte, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	return readLimited(f, maxBytes)
}

// readLimited reads r whole, failing with ErrTooLarge past maxBytes so a
// runaway source cannot exhaust memory.
func readLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(r)
	}

	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("%w: over %d bytes", ErrTooLarge, maxBytes)
	}

	return data, nil
}

// cachePath names a source's cache file by the hash of its URL.
func cachePath(dir, source string) string {
	if dir == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(source))

	return filepath.Join(dir, hex.EncodeToString(sum[:])+".txt")
}

// writeCache stores a good download. Failing to cache only loses the fallback,
// so it is logged rather than returned.
// ponytail: not written atomically; a torn cache file is only ever a fallback
// and still has to clear MIN_RULES before it is published.
func writeCache(ctx context.Context, logger *slog.Logger, path string, data []byte) {
	err := os.MkdirAll(filepath.Dir(path), cacheDirPerm)
	if err == nil {
		err = os.WriteFile(path, data, cacheFilePerm)
	}

	if err != nil {
		logger.LogAttrs(ctx, slog.LevelWarn, "caching blocklist failed", slog.String("error", err.Error()))
	}
}
