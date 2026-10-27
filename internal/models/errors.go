package models

import (
	"errors"
	"fmt"
)

var (
	// ErrMissing reports a required config field that is empty.
	ErrMissing = errors.New("missing")
	// ErrInvalid reports a config field holding an unsupported value.
	ErrInvalid = errors.New("invalid")
	// ErrBadStatus reports a fetch answered with an unexpected HTTP status.
	ErrBadStatus = errors.New("fetch returned status")
	// ErrTooLarge reports a source body over the configured size limit.
	ErrTooLarge = errors.New("source too large")
)

// BlockListError ties a failure to the blocklist it happened in.
type BlockListError struct {
	BlocklistName string
	Err           error
}

func (e BlockListError) Error() string {
	return fmt.Sprintf("blocklist %q: %v", e.BlocklistName, e.Err)
}

func (e BlockListError) Unwrap() error {
	return e.Err
}
