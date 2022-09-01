package app

import (
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
)

type Instance struct {
	baseURL string

	store store.AuthStore
}

func NewInstance(baseURL string, storage store.AuthStore) *Instance {
	return &Instance{
		baseURL: baseURL,
		store:   storage,
	}
}
