package app

import (
	"context"
	"fmt"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
)

func TestNewInstance(t *testing.T) {
	removeCham := make(chan models.BatchRemoveRequest)
	type args struct {
		baseURL    string
		storage    store.AuthStore
		removeChan chan models.BatchRemoveRequest
	}
	var tests = []struct {
		name string
		args args
		want *Instance
	}{
		{
			"test new instance",
			args{baseURL: "http://localhost:8080",
				storage: store.NewInMemory(), removeChan: removeCham},
			NewInstance("http://localhost:8080", store.NewInMemory(), removeCham),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewInstance(tt.args.baseURL, tt.args.storage, tt.args.removeChan), "NewInstance(%v, %v)", tt.args.baseURL, tt.args.storage)
		})
	}
}

func ExampleNewInstance() {
	instance := NewInstance("http://localhost:8080", store.NewInMemory(), make(chan models.BatchRemoveRequest))
	parsedURL, _ := url.Parse("https://praktikum.yandex.ru/")
	id, _ := instance.Store.Save(context.Background(), parsedURL)
	fmt.Println(id)
}
