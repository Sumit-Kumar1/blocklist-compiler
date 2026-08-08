package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const refreshTimeout = 30 * time.Second

// refreshBody asks AdGuard Home to re-read its blocklists rather than its allowlists.
var refreshBody = []byte(`{"whitelist":false}`)

// ErrRefreshFailed reports that AdGuard Home rejected the reload request.
var ErrRefreshFailed = errors.New("adguard refresh returned an unexpected status")

// AdGuard triggers a filter reload on an AdGuard Home instance.
type AdGuard struct {
	api      string
	user     string
	pass     string
	client   *http.Client
	Disabled bool
}

// NewAdGuard returns a client that is Disabled when no API address is configured.
func NewAdGuard(api, user, pass string) *AdGuard {
	return &AdGuard{
		api:      strings.TrimSuffix(api, "/"),
		user:     user,
		pass:     pass,
		client:   &http.Client{Timeout: refreshTimeout},
		Disabled: strings.TrimSpace(api) == "",
	}
}

// Refresh asks AdGuard Home to reload its filters. A failure is not fatal:
// AdGuard re-reads the list on its own update interval regardless.
func (a *AdGuard) Refresh(ctx context.Context) error {
	if a.Disabled {
		return nil
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, a.api+"/control/filtering/refresh", bytes.NewReader(refreshBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if a.user != "" {
		req.SetBasicAuth(a.user, a.pass)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrRefreshFailed, resp.Status)
	}

	return nil
}
