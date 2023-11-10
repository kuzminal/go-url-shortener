package app

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
)

func TestNewInstance(t *testing.T) {
	type args struct {
		baseURL string
		storage store.AuthStore
	}
	var tests = []struct {
		name string
		args args
		want *Instance
	}{
		{
			"test new instance",
			args{baseURL: "http://localhost:8080",
				storage: store.NewInMemory()},
			NewInstance("http://localhost:8080", store.NewInMemory()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewInstance(tt.args.baseURL, tt.args.storage), "NewInstance(%v, %v)", tt.args.baseURL, tt.args.storage)
		})
	}
}

func ExampleNewInstance() {
	instance := NewInstance("http://localhost:8080", store.NewInMemory())
	parsedURL, _ := url.Parse("https://praktikum.yandex.ru/")
	id, _ := instance.store.Save(context.Background(), parsedURL)
	fmt.Println(id)
}
