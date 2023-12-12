package app

import (
	"context"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/auth"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/config"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInstance_Ping(t *testing.T) {
	instance := &Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	t.Run("pint test", func(t *testing.T) {
		instance.Ping(context.Background())
	})
}

func TestInstance_Shorten(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"
	uid := uuid.Must(uuid.NewV4())
	instance := &Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}

	testCases := []struct {
		name          string
		url           string
		expectedError error
		emptyResponse bool
		user          uuid.UUID
	}{
		{
			name:          "bad_request",
			url:           "htt_p://o.com",
			expectedError: ErrParseURL,
			emptyResponse: true,
			user:          uid,
		},
		{
			name:          "success",
			url:           targetURL,
			expectedError: nil,
			emptyResponse: false,
			user:          uid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := auth.Context(context.Background(), tc.user)
			shorten, err := instance.Shorten(ctx, tc.url)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, len(shorten) == 0, tc.emptyResponse)
		})
	}
}

func TestInstance_BatchShorten(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"
	uid := uuid.Must(uuid.NewV4())
	instance := &Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}

	testCases := []struct {
		name          string
		url           string
		expectedError error
		emptyResponse bool
		user          uuid.UUID
	}{
		{
			name:          "bad_request",
			url:           "htt_p://o.com",
			expectedError: ErrParseURL,
			emptyResponse: true,
			user:          uid,
		},
		{
			name:          "success",
			url:           targetURL,
			expectedError: nil,
			emptyResponse: false,
			user:          uid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := auth.Context(context.Background(), tc.user)
			shorten, err := instance.BatchShorten([]models.BatchShortenRequest{{CorrelationID: "1", OriginalURL: tc.url}}, ctx)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, len(shorten) == 0, tc.emptyResponse)
		})
	}
}

func TestInstance_LoadURL(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"
	uid := uuid.Must(uuid.NewV4())
	instance := &Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	ctx := auth.Context(context.Background(), uid)

	testCases := []struct {
		name          string
		id            string
		hasError      bool
		emptyResponse bool
		user          uuid.UUID
	}{
		{
			name:          "bad_request",
			id:            "",
			hasError:      true,
			emptyResponse: true,
			user:          uid,
		},
		{
			name:          "success",
			id:            "0",
			hasError:      false,
			emptyResponse: false,
			user:          uid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.id) > 0 {
				shorten, err := instance.Shorten(ctx, targetURL)
				assert.NoError(t, err)
				assert.NotEmpty(t, shorten)
			}
			shorten, err := instance.LoadURL(ctx, tc.id)
			assert.Equal(t, tc.hasError, err != nil)
			assert.Equal(t, len(shorten.String()) == 0, tc.emptyResponse)
		})
	}
}

func TestInstance_Statistics(t *testing.T) {
	instance := &Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	config.TrustedSubnet = "10.0.0.105/8"
	testCases := []struct {
		name          string
		ip            string
		hasError      bool
		emptyResponse bool
	}{
		{
			name:          "bad_ip",
			ip:            "192.168.1.1",
			hasError:      true,
			emptyResponse: true,
		},
		{
			name:          "success",
			ip:            "10.0.0.105",
			hasError:      false,
			emptyResponse: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := instance.Statistics(context.Background(), tc.ip)
			assert.Equal(t, tc.hasError, err != nil)
		})
	}
}

func TestInstance_LoadUsers(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"
	uid := uuid.Must(uuid.NewV4())
	instance := &Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}

	t.Run("not found", func(t *testing.T) {
		ctx := auth.Context(context.Background(), uid)
		shorten, err := instance.LoadUsers(ctx)
		assert.Error(t, err)
		assert.Empty(t, shorten)
	})
	t.Run("success", func(t *testing.T) {
		ctx := auth.Context(context.Background(), uid)
		_, err := instance.Shorten(ctx, targetURL)
		assert.NoError(t, err)
		shorten, err := instance.LoadUsers(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, shorten)
	})
}
