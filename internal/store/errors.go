package store

import (
	"errors"
)

// Типы пользовательских ошибок
var (
	ErrNotFound = errors.New("not found") // ErrNotFound пользователь или ссылки не найдены
	ErrConflict = errors.New("conflict")  // ErrConflict конфликт обновления\создания записи
)
