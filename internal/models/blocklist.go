package models

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
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
	fetchAttempts  = 3
	fetchTimeout   = 25 * time.Second
	fetchBackoff   = 300 * time.Millisecond
	copyBufferSize = 1 << 20
)

var retryStatuses = []int{
	http.StatusTooEarly, http.StatusTooManyRequests, http.StatusGatewayTimeout,
	http.StatusPartialContent, http.StatusServiceUnavailable,
}

type Blocklist struct {
	Name            string        `json:"name"`
	Source          string        `json:"source"`
	Type            BlocklistType `json:"type"`
	Transformations []string      `json:"transformations"`
	TempName        string
}

func (b Blocklist) validate() error {
	if strings.TrimSpace(b.Source) == "" {
		return ErrMissing("source", "blocklist")
	}

	if strings.TrimSpace(b.Name) == "" {
		return ErrMissing("name", "blocklist")
	}

	if b.Type != hostsType && b.Type != adblockType {
		return ErrInvalid("blocklist type", b.Name)
	}

	return nil
}

// fetch downloads the source, retrying transient failures, and writes the body
// to a temp file under the OS temp dir.
func (b *Blocklist) fetch(ctx context.Context) error {
	logger := slog.With(slog.Any("blocklist", b.Name))

	if ctx.Err() != nil {
		return ErrCtxCancelled("fetching blocklist", b.Name)
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "started fetching blocklist")

	u, err := url.Parse(b.Source)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error while parsing source url", slog.String("error", err.Error()))
		return err
	}

	for attempt := 1; attempt <= fetchAttempts; attempt++ {
		retry, err := b.fetchOnce(ctx, u.String())
		if err == nil {
			return nil
		}

		if !retry || attempt == fetchAttempts {
			logger.LogAttrs(ctx, slog.LevelError, "fetching blocklist failed", slog.String("error", err.Error()))
			return err
		}

		logger.LogAttrs(ctx, slog.LevelWarn, "retrying fetching blocklist", slog.String("error", err.Error()))

		select {
		case <-ctx.Done():
			return ErrCtxCancelled("fetching blocklist", b.Name)
		case <-time.After(time.Duration(attempt) * fetchBackoff):
		}
	}

	return ErrCtxCancelled("fetching blocklist", b.Name)
}

// fetchOnce performs a single download attempt and reports whether the failure
// it returns is worth retrying.
func (b *Blocklist) fetchOnce(ctx context.Context, source string) (bool, error) {
	c, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(c, http.MethodGet, source, http.NoBody)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// This attempt's own timeout is worth another try; a cancelled parent is not.
		return ctx.Err() == nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if slices.Contains(retryStatuses, resp.StatusCode) {
		return true, ErrBadStatus(resp.Status, b.Name)
	}

	if resp.StatusCode != http.StatusOK {
		return false, ErrBadStatus(resp.Status, b.Name)
	}

	return false, b.writeTemp(ctx, resp.Body)
}

// writeTemp streams body into a temp file and records its path on the blocklist.
func (b *Blocklist) writeTemp(ctx context.Context, body io.Reader) (err error) {
	ff, err := os.CreateTemp(os.TempDir(), string(b.Type)+"*.txt")
	if err != nil {
		return err
	}

	b.TempName = ff.Name()

	defer func() {
		if cerr := ff.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if ctx.Err() != nil {
		return ErrCtxCancelled("before writing tempfile", b.Name)
	}

	if _, err := io.CopyBuffer(ff, body, make([]byte, copyBufferSize)); err != nil {
		return err
	}

	return nil
}
