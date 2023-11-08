package app

import (
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
)

// Instance структура для хранения ссылок
type Instance struct {
	baseURL string

	store store.AuthStore
}

// NewInstance создает новый экземпляр структуры для хранения ссылки
func NewInstance(baseURL string, storage store.AuthStore) *Instance {
	return &Instance{
		baseURL: baseURL,
		store:   storage,
	}
}
