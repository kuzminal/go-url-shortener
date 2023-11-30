package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/app"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/auth"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
)

type Handler struct {
	Instance *app.Instance
}

// ShortenHandler обработчик запроса на сокращение ссылки, который принимает в запросе ссылку в виде строки
func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Cannot read request body"))
		return
	}

	u, err := url.Parse(string(b))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Cannot parse given string as URL"))
		return
	}

	shortURL, err := h.Instance.Shorten(r.Context(), u)
	if err != nil && !errors.Is(err, store.ErrConflict) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	status := http.StatusCreated
	if errors.Is(err, store.ErrConflict) {
		status = http.StatusConflict
	}

	w.WriteHeader(status)
	_, _ = w.Write([]byte(shortURL))
}

// ShortenAPIHandler обработчик запроса на сокращение ссылок, который принимает в запросе структуру ShortenRequest
func (h *Handler) ShortenAPIHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request body given"))
		return
	}

	u, err := url.Parse(req.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Cannot parse given string as URL"))
		return
	}

	shortURL, err := h.Instance.Shorten(r.Context(), u)
	if err != nil && !errors.Is(err, store.ErrConflict) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")

	status := http.StatusCreated
	if errors.Is(err, store.ErrConflict) {
		status = http.StatusConflict
	}

	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(models.ShortenResponse{
		Result: shortURL,
	})

	if err != nil {
		fmt.Printf("cannot write response: %s", err)
	}
}

// ExpandHandler обработчик, возвращающий ссылку из хранилища
func (h *Handler) ExpandHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad ID given"))
		return
	}

	target, err := h.Instance.LoadURL(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, store.ErrDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", target.String())
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// UserURLsHandler обработчик, который возвращает все ссылки пользователя.
// Пользователь при этом берется из контекста запроса
func (h *Handler) UserURLsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	urls, err := h.Instance.LoadUsers(ctx)
	if errors.Is(err, app.ErrAuth) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if errors.Is(err, store.ErrNotFound) || len(urls) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(urls)
}

// BatchShortenAPIHandler пакетная обработка запросов на сокращение ссылок.
// Принимает в запросе структуру BatchShortenRequest
func (h *Handler) BatchShortenAPIHandler(w http.ResponseWriter, r *http.Request) {
	var req []models.BatchShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request body given"))
		return
	}

	if len(req) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Empty URLs list given"))
		return
	}

	shorten, err := h.Instance.BatchShorten(req, r.Context())
	if errors.Is(err, app.ErrParseURL) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Cannot parse given string as URL"))
		return
	}

	if errors.Is(err, app.ErrURLLength) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("invalid shorten URLs length"))
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(shorten)
	if err != nil {
		fmt.Printf("cannot write response: %s", err)
	}
}

// BatchRemoveAPIHandler пакетное удаление пользовательских ссылок ссылок.
// Пользователь определяется из контекста запроса.
func (h *Handler) BatchRemoveAPIHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := auth.UIDFromContext(ctx)
	if uid == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var ids []string
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request body given"))
		return
	}

	if len(ids) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Empty IDs list given"))
		return
	}

	go func() {
		h.Instance.RemoveChan <- models.BatchRemoveRequest{UID: *uid, Ids: ids}
	}()

	w.WriteHeader(http.StatusAccepted)
}

// PingHandler проверяет, что приложение в состоянии обработать запросы
func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	// ensure everything is okay
	err := h.Instance.Ping(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// StatisticsHandler выдает статистику по пользователям и по ссылкам
func (h *Handler) StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("X-Real-IP")
	stat, err := h.Instance.Statistics(r.Context(), ip)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(stat)
	if err != nil {
		fmt.Printf("cannot write response: %s", err)
	}
}
