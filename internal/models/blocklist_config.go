package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type BlocklistConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
	License     string `json:"license"`
	Version     string `json:"version"`

	Blocklists      []Blocklist `json:"sources"`
	Transformations []string    `json:"transformations"`
	Filters
}

// Filters drop or keep rules by pattern, mirroring @adguard/hostlist-compiler.
// The *_sources fields name URLs or files holding one pattern per line, and
// Process appends their contents to Exclusions and Inclusions.
type Filters struct {
	Exclusions        []string `json:"exclusions"`
	ExclusionsSources []string `json:"exclusions_sources"`
	Inclusions        []string `json:"inclusions"`
	InclusionsSources []string `json:"inclusions_sources"`
}

func (c BlocklistConfig) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("%w: name in blocklist config", ErrMissing)
	}

	if len(c.Blocklists) == 0 {
		return fmt.Errorf("%w: sources in blocklist config", ErrMissing)
	}

	if len(c.Transformations) == 0 {
		return fmt.Errorf("%w: transformations in blocklist config", ErrMissing)
	}

	for _, blocklist := range c.Blocklists {
		if err := blocklist.validate(); err != nil {
			return err
		}
	}

	return nil
}

// Process fetches every blocklist and every pattern source concurrently and
// reports all failures.
func (c *BlocklistConfig) Process(ctx context.Context, opts FetchOptions) error {
	errs := make([]error, len(c.Blocklists)+1)

	var wg sync.WaitGroup

	wg.Go(func() {
		if err := c.load(ctx, opts); err != nil {
			errs[0] = fmt.Errorf("blocklist config filters: %w", err)
		}
	})

	for i := range c.Blocklists {
		blocklist := &c.Blocklists[i]

		wg.Go(func() {
			err := blocklist.fetch(ctx, opts)
			if err == nil {
				err = blocklist.load(ctx, opts)
			}

			if err != nil {
				errs[i+1] = BlockListError{BlocklistName: blocklist.Name, Err: err}
			}
		})
	}

	wg.Wait()

	return errors.Join(errs...)
}

// load appends the patterns from the *_sources fields.
func (f *Filters) load(ctx context.Context, opts FetchOptions) error {
	var err error

	if f.Exclusions, err = appendPatterns(ctx, opts, f.Exclusions, f.ExclusionsSources); err != nil {
		return err
	}

	f.Inclusions, err = appendPatterns(ctx, opts, f.Inclusions, f.InclusionsSources)

	return err
}

func appendPatterns(ctx context.Context, opts FetchOptions, patterns, sources []string) ([]string, error) {
	for _, source := range sources {
		src := Blocklist{Name: source, Source: source}

		if err := src.fetch(ctx, opts); err != nil {
			return nil, fmt.Errorf("loading patterns from %s: %w", source, err)
		}

		patterns = append(patterns, strings.Split(string(src.Data), "\n")...)
	}

	return patterns, nil
}
