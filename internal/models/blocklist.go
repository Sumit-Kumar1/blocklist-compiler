package models

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type BlocklistType string

const (
	adblockType BlocklistType = "adblock"
	hostsType   BlocklistType = "hosts"
)

type Blocklist struct {
	Name            string        `json:"name"`
	Source          string        `json:"source"`
	Type            BlocklistType `json:"type"`
	Transformations []string      `json:"transformations"`
	TempName        string
}

func (b Blocklist) validate() error {
	if strings.TrimSpace(b.Source) == "" {
		return errors.New("no source found in blocklist")
	}

	if strings.TrimSpace(b.Name) == "" {
		return errors.New("no name found in blocklist")
	}

	if b.Type != hostsType && b.Type != adblockType {
		return errors.New("invalid blocklist type, supported only : adblock, hosts")
	}

	return nil
}

// process checks each sources, along with fetch and write each source under blc temp dir
func (b *Blocklist) process() error {
	u, err := url.Parse(b.Source)
	if err != nil {
		return err
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}

	if resp == nil {
		return errors.New("recieved empty response")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response statuscode : %d", resp.StatusCode)
	}

	ff, err := os.CreateTemp("blc", string(b.Type)+"*.txt")
	if err != nil {
		return err
	}

	b.TempName = ff.Name()

	defer ff.Close()
	defer resp.Body.Close()

	_, err = io.CopyBuffer(ff, resp.Body, make([]byte, 1024))
	if err != nil {
		return err
	}

	return nil
}
