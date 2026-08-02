package internal

import (
	"errors"
	"strings"
)

type blocklistType string

const (
	adblockType blocklistType = "adblock"
	hostsType   blocklistType = "hosts"
)

type blocklist struct {
	Name            string        `json:"name"`
	Source          string        `json:"source"`
	Type            blocklistType `json:"type"`
	Transformations []string      `json:"transformations"`
}

func (b blocklist) validate() error {
	if strings.TrimSpace(b.Source) == "" {
		return errors.New("no source found in blocklist")
	}

	return nil
}

func (b blocklist) process() error {
	return nil
}
