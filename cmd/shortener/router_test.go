package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/auth"
)

func Test_authMiddleware(t *testing.T) {
	t.Run("no_cookie", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/user/urls", nil)
		w := httptest.NewRecorder()

		mw := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotNil(t, auth.UIDFromContext(r.Context()))
		}))
		mw.ServeHTTP(w, r)

		assert.NotEmpty(t, w.Header().Get("Set-Cookie"))
	})

	t.Run("bad_cookie", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/user/urls", nil)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "ololo"})
		w := httptest.NewRecorder()

		mw := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotNil(t, auth.UIDFromContext(r.Context()))
		}))
		mw.ServeHTTP(w, r)

		assert.NotEmpty(t, w.Header().Get("Set-Cookie"))
	})

	t.Run("existing_cookie", func(t *testing.T) {
		uid := uuid.Must(uuid.NewV4())
		cookie, err := auth.EncodeUIDToHex(uid)
		require.NoError(t, err)

		r := httptest.NewRequest("GET", "/api/user/urls", nil)
		r.AddCookie(&http.Cookie{Name: "auth", Value: cookie})
		w := httptest.NewRecorder()

		mw := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, &uid, auth.UIDFromContext(r.Context()))
		}))
		mw.ServeHTTP(w, r)

		assert.Empty(t, w.Header().Get("Set-Cookie"))
	})
}
