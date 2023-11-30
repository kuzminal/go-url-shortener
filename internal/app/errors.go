package app

import "errors"

var (
	ErrAuth      = errors.New("auth unprocessed")
	ErrParseUrl  = errors.New("cannot parse given string as URL")
	ErrUrlLength = errors.New("invalid shorten URLs length")
)
