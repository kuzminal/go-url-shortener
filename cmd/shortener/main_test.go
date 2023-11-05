package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/config"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
)

func TestMain(m *testing.M) {
	config.Parse()
	go func() {
		err := run()
		if err != nil {
			panic(err)
		}
	}()
	os.Exit(m.Run())
}

func Test_run(t *testing.T) {
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

				defer func(Body io.ReadCloser) {
					_ = Body.Close()
				}(resp.Body)

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

				defer func(Body io.ReadCloser) {
					_ = Body.Close()
				}(resp.Body)

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

				defer func(Body io.ReadCloser) {
					_ = Body.Close()
				}(resp.Body)

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

				defer func(Body io.ReadCloser) {
					_ = Body.Close()
				}(resp.Body)

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

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

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

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		expectResponse := "http://localhost:8080/65"
		actualResponse := string(b)

		require.Equal(t, expectResponse, actualResponse)
	})
}

func TestEndToEnd(t *testing.T) {
	originalURL := "https://praktikum.yandex.ru"

	// create HTTP client without redirects support
	errRedirectBlocked := errors.New("HTTP redirect blocked")
	redirPolicy := resty.RedirectPolicyFunc(func(_ *http.Request, _ []*http.Request) error {
		return errRedirectBlocked
	})

	httpc := resty.New().
		SetBaseURL("http://localhost:8080").
		SetRedirectPolicy(redirPolicy)

	// shorten URL
	req := httpc.R().
		SetBody(originalURL)
	resp, err := req.Post("/")
	require.NoError(t, err)

	shortenURL := string(resp.Body())

	assert.Equal(t, http.StatusCreated, resp.StatusCode())
	assert.NoError(t, func() error {
		_, errs := url.Parse(shortenURL)
		return errs
	}())

	// expand URL
	req = resty.New().
		SetRedirectPolicy(redirPolicy).
		R()
	resp, err = req.Get(shortenURL)
	if !errors.Is(err, errRedirectBlocked) {
		require.NoError(t, err)
	}

	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode())
	assert.Equal(t, originalURL, resp.Header().Get("Location"))
}

func Test_newStore(t *testing.T) {
	t.Run("create file store", func(t *testing.T) {
		config.PersistFile = "./store"
		store, err := newStore(context.Background())
		require.NoError(t, err)
		require.NotNil(t, store)
	})

	t.Run("create pg store", func(t *testing.T) {
		config.DatabaseDSN = "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
		store, err := newStore(context.Background())
		require.NoError(t, err)
		require.NotNil(t, store)
	})
}
