package app

import (
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
)

// Instance структура для хранения информации о приложении, включая хранилище
type Instance struct {
	baseURL string

	store      store.AuthStore
	removeChan chan models.BatchRemoveRequest
}

// NewInstance функция создания новой структуры приложения
func NewInstance(baseURL string, storage store.AuthStore, removeChan chan models.BatchRemoveRequest) *Instance {
	return &Instance{
		baseURL:    baseURL,
		store:      storage,
		removeChan: removeChan,
	}
}
