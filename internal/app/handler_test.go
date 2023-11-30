package app

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/auth"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
)

func Test_ShortenAPIHandler(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"

	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   store.NewInMemory(),
	}

	testCases := []struct {
		name             string
		url              string
		expectedStatus   int
		expectedResponse []byte
	}{
		{
			name:             "bad_request",
			url:              "htt_p://o.com",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: []byte("Cannot parse given string as URL"),
		},
		{
			name:             "success",
			url:              targetURL,
			expectedStatus:   http.StatusCreated,
			expectedResponse: []byte("{\"result\":\"http://localhost:8080/0\"}\n"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(models.ShortenRequest{URL: tc.url})
			require.NoError(t, err)
			body := bytes.NewBuffer(b)

			r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten", body)
			w := httptest.NewRecorder()

			instance.ShortenAPIHandler(w, r)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedResponse, w.Body.Bytes())
		})
	}
}

func Test_expander(t *testing.T) {
	expectedURL := "https://praktikum.yandex.ru/"
	parsedURL, _ := url.Parse(expectedURL)

	storage := store.NewInMemory()
	id, _ := storage.Save(context.Background(), parsedURL)

	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   storage,
	}

	testCases := []struct {
		name             string
		id               string
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "bad_request",
			id:               "",
			expectedStatus:   http.StatusBadRequest,
			expectedLocation: "",
		},
		{
			name:             "not_found",
			id:               "-1",
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
		{
			name:             "success",
			id:               id,
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: expectedURL,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "http://localhost:8080/"+tc.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			instance.ExpandHandler(w, r)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedLocation, w.Header().Get("Location"))
		})
	}
}

func Test_userURLs(t *testing.T) {
	uid := uuid.Must(uuid.NewV4())
	u, _ := url.Parse("https://praktikum.yandex.ru/")

	storage := store.NewInMemory()
	id, _ := storage.SaveUser(context.Background(), uid, u)

	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   storage,
	}

	testCases := []struct {
		name           string
		ctx            context.Context
		expectedStatus int
		expectedBody   []byte
	}{
		{
			name:           "no_uid",
			ctx:            context.Background(),
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   nil,
		},
		{
			name:           "no_urls",
			ctx:            auth.Context(context.Background(), uuid.Must(uuid.NewV4())),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   nil,
		},
		{
			name:           "has_urls",
			ctx:            auth.Context(context.Background(), uid),
			expectedStatus: http.StatusOK,
			expectedBody:   []byte("[{\"short_url\":\"http://localhost:8080/" + id + "\",\"original_url\":\"https://praktikum.yandex.ru/\"}]\n"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "http://localhost:8080/user/urls", nil)
			r = r.WithContext(tc.ctx)

			w := httptest.NewRecorder()
			instance.UserURLsHandler(w, r)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedBody, w.Body.Bytes())
		})
	}
}

func Test_ShortenHandler(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"

	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   store.NewInMemory(),
	}

	testCases := []struct {
		name             string
		url              string
		expectedStatus   int
		expectedResponse []byte
	}{
		{
			name:             "bad_request",
			url:              "htt_p://o.com",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: []byte("Cannot parse given string as URL"),
		},
		{
			name:             "success",
			url:              targetURL,
			expectedStatus:   http.StatusCreated,
			expectedResponse: []byte("http://localhost:8080/0"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := bytes.NewBuffer([]byte(tc.url))

			r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten", body)
			w := httptest.NewRecorder()

			instance.ShortenHandler(w, r)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedResponse, w.Body.Bytes())
		})
	}
}

func TestInstance_PingHandler(t *testing.T) {
	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   store.NewInMemory(),
	}
	t.Run("ping", func(t *testing.T) {
		r := httptest.NewRequest("GET", "http://localhost:8080/ping", nil)
		w := httptest.NewRecorder()

		instance.PingHandler(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func ExampleInstance_ShortenHandler() {
	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   store.NewInMemory(),
	}

	b := "https://praktikum.yandex.ru/"
	body := bytes.NewBuffer([]byte(b))

	r := httptest.NewRequest("POST", "http://localhost:8080/", body)
	w := httptest.NewRecorder()
	instance.ShortenHandler(w, r)
}

func ExampleInstance_ShortenAPIHandler() {
	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   store.NewInMemory(),
	}

	b, _ := json.Marshal(models.ShortenRequest{URL: "https://praktikum.yandex.ru/"})
	body := bytes.NewBuffer(b)

	r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten", body)
	w := httptest.NewRecorder()
	instance.ShortenHandler(w, r)
}

func ExampleInstance_BatchShortenAPIHandler() {
	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   store.NewInMemory(),
	}

	b, _ := json.Marshal(models.BatchShortenRequest{CorrelationID: "id1", OriginalURL: "https://praktikum.yandex.ru/"})
	body := bytes.NewBuffer(b)

	r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten", body)
	w := httptest.NewRecorder()
	instance.BatchShortenAPIHandler(w, r)
}

func ExampleInstance_BatchRemoveAPIHandler() {
	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   store.NewInMemory(),
	}

	ids := []string{"id1", "id2", "id3"}
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(ids)
	body := buf

	r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten", body)
	w := httptest.NewRecorder()
	instance.BatchRemoveAPIHandler(w, r)
}

func ExampleInstance_PingHandler() {
	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   store.NewInMemory(),
	}
	r := httptest.NewRequest("GET", "http://localhost:8080/ping", nil)
	w := httptest.NewRecorder()

	instance.PingHandler(w, r)
}

func ExampleInstance_ExpandHandler() {
	expectedURL := "https://praktikum.yandex.ru/"
	parsedURL, _ := url.Parse(expectedURL)

	storage := store.NewInMemory()
	id, _ := storage.Save(context.Background(), parsedURL)

	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   storage,
	}

	r := httptest.NewRequest("GET", "http://localhost:8080/"+id, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	instance.ExpandHandler(w, r)
}

func ExampleInstance_UserURLsHandler() {
	uid := uuid.Must(uuid.NewV4())
	storage := store.NewInMemory()
	u, _ := url.Parse("https://praktikum.yandex.ru/")
	storage.SaveUser(context.Background(), uid, u)

	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   storage,
	}

	r := httptest.NewRequest("GET", "http://localhost:8080/user/urls", nil)
	r = r.WithContext(auth.Context(context.Background(), uid))

	w := httptest.NewRecorder()
	instance.UserURLsHandler(w, r)
}

func TestInstance_BatchShortenAPIHandler(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"

	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   store.NewInMemory(),
	}

	testCases := []struct {
		name             string
		url              string
		expectedStatus   int
		expectedResponse []byte
	}{
		{
			name:             "bad_request",
			url:              "htt_p://o.com",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: []byte("Cannot parse given string as URL: htt_p://o.com"),
		},
		{
			name:             "success",
			url:              targetURL,
			expectedStatus:   http.StatusCreated,
			expectedResponse: []byte("[{\"correlation_id\":\"1\",\"short_url\":\"http://localhost:8080/0\"}]\n"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal([]models.BatchShortenRequest{{CorrelationID: "1", OriginalURL: tc.url}})
			require.NoError(t, err)
			body := bytes.NewBuffer(b)

			r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten/batch", body)
			w := httptest.NewRecorder()

			instance.BatchShortenAPIHandler(w, r)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedResponse, w.Body.Bytes())
		})
	}
}

func TestInstance_BatchRemoveAPIHandler(t *testing.T) {
	uid := uuid.Must(uuid.NewV4())
	storage := store.NewInMemory()
	baseURL := "https://praktikum.yandex.ru/"
	var urls []string
	for i := 1; i < 100; i++ {
		tempURL := baseURL + strconv.Itoa(i)
		u, _ := url.Parse(tempURL)
		storage.SaveUser(context.Background(), uid, u)
		urls = append(urls, tempURL)
	}

	instance := &Instance{
		baseURL: "http://localhost:8080",
		store:   storage,
	}
	b, _ := json.Marshal(urls)
	body := bytes.NewBuffer(b)
	r := httptest.NewRequest("DELETE", "http://localhost:8080/api/user/urls", body)
	r = r.WithContext(auth.Context(context.Background(), uid))

	w := httptest.NewRecorder()
	instance.BatchRemoveAPIHandler(w, r)
	log.Println(w.Body)
	assert.Equal(t, http.StatusAccepted, w.Code)
}
