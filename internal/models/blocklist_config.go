package models

import (
	"errors"
	"fmt"
	"strings"
)

type BlocklistConfig struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Blocklists      []Blocklist `json:"sources"`
	Transformations []string    `json:"transformations"`
	Exclusions      []string    `json:"exclusions"`
}

func (c BlocklistConfig) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("missing name in blocklist config")
	}

	if len(c.Blocklists) == 0 {
		return errors.New("missing sources in blocklist config")
	}

	if len(c.Transformations) == 0 {
		return errors.New("missing transformations in blocklist config")
	}

	for _, blocklist := range c.Blocklists {
		if err := blocklist.validate(); err != nil {
			return err
		}
	}

	return nil
}

// processBlocklists fetches each blocklists using blocklist.Process
func (c *BlocklistConfig) Process() error {

	for _, blocklist := range c.Blocklists {
		if err := blocklist.process(); err != nil {
			return fmt.Errorf("error while processing blocklist: %s, source: %s", blocklist.Name, blocklist.Source)
		}
	}

	return nil
}
