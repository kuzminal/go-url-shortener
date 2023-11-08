package store_test

import (
	"context"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
	"net/url"
)

func ExampleNewInMemory() {
	ctx := context.Background()
	storage := store.NewInMemory()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	_, _ = storage.Save(ctx, urlToStore)
}
