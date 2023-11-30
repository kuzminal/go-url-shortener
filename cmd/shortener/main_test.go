package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/config"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pgContainer *postgres.PostgresContainer
)

func CreatePostgresContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15.3-alpine"),
		//postgres.WithInitScripts(filepath.Join("..", "testdata", "init-db.sql")),
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	return pgContainer, err
}

func Test_run(t *testing.T) {
	config.ShutdownTimeout = 10 * time.Minute
	cancelableContext, cancel := context.WithCancel(context.Background())
	go func() {
		err := run(cancelableContext)
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(200 * time.Millisecond)

	targetURL := "https://praktikum.yandex.ru/"

	for i := 0; i < 50; i++ {
		expectedID := fmt.Sprintf("%x", i)

		t.Run("shorten", func(t *testing.T) {
			expectResponse := "http://localhost:8080/" + expectedID
			var actualResponse string

			{
				body := bytes.NewBufferString(targetURL)
				r := httptest.NewRequest("POST", "http://localhost:8080/", body)
				r.RequestURI = ""

				resp, err := http.DefaultClient.Do(r)
				require.NoError(t, err)
				require.Equal(t, http.StatusCreated, resp.StatusCode)

				defer resp.Body.Close()

				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				actualResponse = string(b)

				require.Equal(t, expectResponse, actualResponse)
			}

			{
				r := httptest.NewRequest("GET", actualResponse, nil)
				r.RequestURI = ""

				resp, err := http.DefaultTransport.RoundTrip(r)
				require.NoError(t, err)

				defer resp.Body.Close()

				assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
				assert.Equal(t, targetURL, resp.Header.Get("Location"))
			}
		})
	}

	for i := 50; i < 100; i++ {
		expectedID := fmt.Sprintf("%x", i)

		t.Run("shortenAPI", func(t *testing.T) {
			expectResponse := "{\"result\":\"http://localhost:8080/" + expectedID + "\"}\n"
			var actualResponse string

			{
				body := bytes.NewBufferString(`{"url":"` + targetURL + `"}`)
				r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten", body)
				r.RequestURI = ""

				resp, err := http.DefaultClient.Do(r)
				require.NoError(t, err)
				require.Equal(t, http.StatusCreated, resp.StatusCode)

				defer resp.Body.Close()

				b, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				actualResponse = string(b)

				require.Equal(t, expectResponse, actualResponse)
			}

			{
				var target models.ShortenResponse
				err := json.Unmarshal([]byte(actualResponse), &target)
				require.NoError(t, err)

				r := httptest.NewRequest("GET", target.Result, nil)
				r.RequestURI = ""

				resp, err := http.DefaultTransport.RoundTrip(r)
				require.NoError(t, err)

				defer resp.Body.Close()

				assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
				assert.Equal(t, targetURL, resp.Header.Get("Location"))
			}
		})
	}

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(targetURL))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", "http://localhost:8080/", buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		expectResponse := "http://localhost:8080/64"
		actualResponse := string(b)

		require.Equal(t, expectResponse, actualResponse)
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(targetURL)
		r := httptest.NewRequest("POST", "http://localhost:8080/", buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		expectResponse := "http://localhost:8080/65"
		actualResponse := string(b)

		require.Equal(t, expectResponse, actualResponse)
	})
	cancel()
	time.Sleep(200 * time.Millisecond)

	timeOutContext, cancelWithTimeOut := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancelWithTimeOut()
	t.Run("graceful_shutdown_exceeded", func(t *testing.T) {
		config.ShutdownTimeout = 1 * time.Nanosecond
		err := run(timeOutContext)
		require.Error(t, err)
		require.EqualError(t, err, fmt.Sprintf("server shutdown: %s", context.DeadlineExceeded.Error()))
		time.Sleep(100 * time.Millisecond)
	})
	t.Run("graceful_shutdown_normal", func(t *testing.T) {
		config.ShutdownTimeout = 1 * time.Minute
		err := run(timeOutContext)
		require.NoError(t, err)
	})
}

func Test_newStore(t *testing.T) {
	t.Run("create file store", func(t *testing.T) {
		config.PersistFile = "./store"
		store, err := newStore(context.Background())
		require.NoError(t, err)
		require.NotNil(t, store)
	})
	pgContainer, _ = CreatePostgresContainer(context.Background())

	t.Run("create pg store", func(t *testing.T) {
		connStr, _ := pgContainer.ConnectionString(context.Background(), "sslmode=disable")
		config.DatabaseDSN = connStr
		store, err := newStore(context.Background())
		require.NoError(t, err)
		require.NotNil(t, store)
	})
	pgContainer.Terminate(context.Background())
}
