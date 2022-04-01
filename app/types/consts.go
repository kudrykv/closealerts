package types

import "errors"

var (
	ErrLinkExists      = errors.New("link exists")
	ErrUnknownCBAction = errors.New("unknown action")
)
