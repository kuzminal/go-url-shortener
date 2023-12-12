package http

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/app"
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
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/config"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
)

func Test_ShortenAPIHandler(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"

	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	handler := Handler{Instance: instance}

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

			handler.ShortenAPIHandler(w, r)

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

	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   storage,
	}
	handler := Handler{Instance: instance}

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

			handler.ExpandHandler(w, r)

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

	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   storage,
	}
	handler := Handler{Instance: instance}

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
			handler.UserURLsHandler(w, r)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedBody, w.Body.Bytes())
		})
	}
}

func Test_ShortenHandler(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"

	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	handler := Handler{Instance: instance}

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

			handler.ShortenHandler(w, r)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedResponse, w.Body.Bytes())
		})
	}
}

func TestInstance_PingHandler(t *testing.T) {
	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	handler := Handler{Instance: instance}
	t.Run("ping", func(t *testing.T) {
		r := httptest.NewRequest("GET", "http://localhost:8080/ping", nil)
		w := httptest.NewRecorder()

		handler.PingHandler(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func ExampleHandler_ShortenHandler() {
	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	handler := Handler{Instance: instance}

	b := "https://praktikum.yandex.ru/"
	body := bytes.NewBuffer([]byte(b))

	r := httptest.NewRequest("POST", "http://localhost:8080/", body)
	w := httptest.NewRecorder()
	handler.ShortenHandler(w, r)
}

func ExampleHandler_ShortenAPIHandler() {
	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	handler := Handler{Instance: instance}

	b, _ := json.Marshal(models.ShortenRequest{URL: "https://praktikum.yandex.ru/"})
	body := bytes.NewBuffer(b)

	r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten", body)
	w := httptest.NewRecorder()
	handler.ShortenHandler(w, r)
}

func ExampleHandler_BatchShortenAPIHandler() {
	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	handler := Handler{Instance: instance}

	b, _ := json.Marshal(models.BatchShortenRequest{CorrelationID: "id1", OriginalURL: "https://praktikum.yandex.ru/"})
	body := bytes.NewBuffer(b)

	r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten", body)
	w := httptest.NewRecorder()
	handler.BatchShortenAPIHandler(w, r)
}

func ExampleHandler_BatchRemoveAPIHandler() {
	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	handler := Handler{Instance: instance}

	ids := []string{"id1", "id2", "id3"}
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(ids)
	body := buf

	r := httptest.NewRequest("POST", "http://localhost:8080/api/shorten", body)
	w := httptest.NewRecorder()
	handler.BatchRemoveAPIHandler(w, r)
}

func ExampleHandler_PingHandler() {
	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	handler := Handler{Instance: instance}
	r := httptest.NewRequest("GET", "http://localhost:8080/ping", nil)
	w := httptest.NewRecorder()

	handler.PingHandler(w, r)
}

func ExampleHandler_ExpandHandler() {
	expectedURL := "https://praktikum.yandex.ru/"
	parsedURL, _ := url.Parse(expectedURL)

	storage := store.NewInMemory()
	id, _ := storage.Save(context.Background(), parsedURL)

	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   storage,
	}
	handler := Handler{Instance: instance}

	r := httptest.NewRequest("GET", "http://localhost:8080/"+id, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ExpandHandler(w, r)
}

func ExampleHandler_UserURLsHandler() {
	uid := uuid.Must(uuid.NewV4())
	storage := store.NewInMemory()
	u, _ := url.Parse("https://praktikum.yandex.ru/")
	storage.SaveUser(context.Background(), uid, u)

	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   storage,
	}
	handler := Handler{Instance: instance}

	r := httptest.NewRequest("GET", "http://localhost:8080/user/urls", nil)
	r = r.WithContext(auth.Context(context.Background(), uid))

	w := httptest.NewRecorder()
	handler.UserURLsHandler(w, r)
}

func TestInstance_BatchShortenAPIHandler(t *testing.T) {
	targetURL := "https://praktikum.yandex.ru/"

	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   store.NewInMemory(),
	}
	handler := Handler{Instance: instance}

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

			handler.BatchShortenAPIHandler(w, r)

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

	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   storage,
	}
	handler := Handler{Instance: instance}
	b, _ := json.Marshal(urls)
	body := bytes.NewBuffer(b)
	r := httptest.NewRequest("DELETE", "http://localhost:8080/api/user/urls", body)
	r = r.WithContext(auth.Context(context.Background(), uid))

	w := httptest.NewRecorder()
	handler.BatchRemoveAPIHandler(w, r)
	log.Println(w.Body)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestInstance_StatisticsHandler(t *testing.T) {
	uid := uuid.Must(uuid.NewV4())
	storage := store.NewInMemory()
	baseURL := "https://praktikum.yandex.ru/"

	for i := 1; i < 100; i++ {
		tempURL := baseURL + strconv.Itoa(i)
		u, _ := url.Parse(tempURL)
		storage.SaveUser(context.Background(), uid, u)
	}
	instance := &app.Instance{
		BaseURL: "http://localhost:8080",
		Store:   storage,
	}
	handler := Handler{Instance: instance}

	testCases := []struct {
		name             string
		instance         *app.Instance
		ip               string
		trustedSubnet    string
		expectedStatus   int
		expectedResponse models.Statistics
	}{
		{
			name:             "forbiden",
			instance:         instance,
			ip:               "192.168.1.123",
			trustedSubnet:    "10.0.0.105/8",
			expectedStatus:   http.StatusForbidden,
			expectedResponse: models.Statistics{},
		},
		{
			name:             "ok",
			instance:         instance,
			ip:               "10.0.0.105",
			trustedSubnet:    "10.0.0.105/8",
			expectedStatus:   http.StatusOK,
			expectedResponse: models.Statistics{Users: 1, Urls: 99},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "http://localhost:8080/api/internal/stats", nil)
			r = r.WithContext(auth.Context(context.Background(), uid))
			config.TrustedSubnet = tc.trustedSubnet
			r.Header.Add("X-Real-IP", tc.ip)
			w := httptest.NewRecorder()
			handler.StatisticsHandler(w, r)
			var resp models.Statistics
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedResponse, resp)
		})
	}
}
