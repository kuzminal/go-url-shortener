package app

import (
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
)

// Instance структура для хранения информации о приложении, включая хранилище
type Instance struct {
	baseURL string

	store store.AuthStore
}

// NewInstance функция создания новой структуры приложения
func NewInstance(baseURL string, storage store.AuthStore) *Instance {
	return &Instance{
		baseURL: baseURL,
		store:   storage,
	}
}
