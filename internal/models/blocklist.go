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

// fetch checks each sources, along with fetch and write each source under blc temp dir
func (b *Blocklist) fetch(ctx context.Context) error {
	logger := slog.With(slog.Any("blocklist", b.Name))

	if ctx.Err() != nil {
		return ErrCtxCancalled("fetching blocklist", b.Name)
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "started fetching blocklist")

	u, err := url.Parse(b.Source)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error while parsing source url", slog.String("error", err.Error()))
		return err
	}

	var resp *http.Response

	sleepTime := 300 * time.Millisecond

	for retry := 1; retry <= 3; retry++ {
		sleepTime *= time.Duration(retry)

		c, cancel := context.WithTimeout(context.TODO(), 25*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(c, http.MethodGet, u.String(), http.NoBody)
		if err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "error while creating new fetch request", slog.String("error", err.Error()))
			return err
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "error while making request", slog.String("error", err.Error()))
			return err
		}

		if resp == nil {
			return ErrInvalid("fetch http response", b.Name)
		}

		if c.Err() != nil { // context is cancelled due to long network call, retry after a second
			time.Sleep(sleepTime)
			continue
		}

		if !slices.Contains(retryStatuses, resp.StatusCode) {
			break
		}

		time.Sleep(sleepTime)
		logger.LogAttrs(ctx, slog.LevelWarn, "retrying fetching blocklist")
	}

	if ctx.Err() != nil {
		return ErrCtxCancalled("after fetching blocklist", b.Name)
	}

	if resp == nil {
		return BlockListError{Reason: "empty http response", BlocklistName: b.Name}
	}

	defer resp.Body.Close()

	ff, err := os.CreateTemp(os.TempDir(), string(b.Type)+"*.txt")
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error while creating temp file", slog.String("error", err.Error()))
		return err
	}

	b.TempName = ff.Name()

	defer ff.Close()

	if ctx.Err() != nil {
		return ErrCtxCancalled("before wiriting tempfile", b.Name)
	}

	_, err = io.CopyBuffer(ff, resp.Body, make([]byte, 1<<20)) // with 1 MB buffer
	if err != nil {
		return err
	}

	return nil
}
