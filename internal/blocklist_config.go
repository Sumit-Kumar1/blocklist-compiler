package internal

import (
	"errors"
	"strings"
)

type blocklistConfig struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Blocklists      []blocklist `json:"sources"`
	Transformations []string    `json:"transformations"`
	Exclusions      []string    `json:"exclusions"`
}

func (c blocklistConfig) validate() error {
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

func (c blocklistConfig) process() error {
	return nil
}
