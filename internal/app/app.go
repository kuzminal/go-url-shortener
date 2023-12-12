package app

import (
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
)

// Instance структура для хранения информации о приложении, включая хранилище
type Instance struct {
	BaseURL string

	Store      store.AuthStore
	RemoveChan chan models.BatchRemoveRequest
}

// NewInstance функция создания новой структуры приложения
func NewInstance(baseURL string, storage store.AuthStore, removeChan chan models.BatchRemoveRequest) *Instance {
	return &Instance{
		BaseURL:    baseURL,
		Store:      storage,
		RemoveChan: removeChan,
	}
}
