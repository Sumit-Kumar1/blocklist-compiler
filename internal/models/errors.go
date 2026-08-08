package models

import "fmt"

type BlockListError struct {
	Reason        string
	BlocklistName string
}

func (e BlockListError) Error() string {
	return e.Reason
}

func ErrMissing(entity, blocklist string) BlockListError {
	return BlockListError{
		Reason:        fmt.Sprintf("missing %s from blocklist %s", entity, blocklist),
		BlocklistName: blocklist,
	}
}

func ErrInvalid(entity, blocklist string) BlockListError {
	return BlockListError{
		Reason:        fmt.Sprintf("invalid %s entity in blocklist %s", entity, blocklist),
		BlocklistName: blocklist,
	}
}

func ErrCtxCancalled(stage, blocklist string) BlockListError {
	return BlockListError{
		Reason:        "context is cancelled at " + stage,
		BlocklistName: blocklist,
	}
}
