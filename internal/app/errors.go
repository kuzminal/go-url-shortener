package app

import "errors"

var (
	ErrAuth      = errors.New("auth unprocessed")                 // ErrAuth ошибка авторизации
	ErrParseURL  = errors.New("cannot parse given string as URL") // ErrParseURL ошибка парсинга строки в URL
	ErrURLLength = errors.New("invalid shorten URLs length")      //ErrURLLength ошибка длины ссылки
)
