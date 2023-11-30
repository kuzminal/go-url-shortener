package app

import "errors"

var (
	ErrAuth      = errors.New("auth unprocessed")
	ErrParseURL  = errors.New("cannot parse given string as URL")
	ErrURLLength = errors.New("invalid shorten URLs length")
)
