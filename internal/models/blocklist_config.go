package models

import (
	"context"
	"strings"
)

const blocklistCfg = "blocklist config"

type BlocklistConfig struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Blocklists      []Blocklist `json:"sources"`
	Transformations []string    `json:"transformations"`
	Exclusions      []string    `json:"exclusions"`
}

func (c BlocklistConfig) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return ErrMissing("name", blocklistCfg)
	}

	if len(c.Blocklists) == 0 {
		return ErrMissing("sources", blocklistCfg)
	}

	if len(c.Transformations) == 0 {
		return ErrMissing("transformations", blocklistCfg)
	}

	for _, blocklist := range c.Blocklists {
		if err := blocklist.validate(); err != nil {
			return err
		}
	}

	return nil
}

// processBlocklists fetches each blocklists using blocklist.Process
func (c *BlocklistConfig) Process(ctx context.Context) error {
	for i := range c.Blocklists {
		if ctx.Err() != nil {
			return ErrCtxCancelled("process blocklist-config", c.Name)
		}

		if err := c.Blocklists[i].fetch(ctx); err != nil {
			return BlockListError{Reason: err.Error(), BlocklistName: c.Blocklists[i].Name}
		}
	}

	return nil
}
