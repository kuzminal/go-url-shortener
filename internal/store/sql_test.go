package store

import (
	"context"
	"database/sql"
	"github.com/gofrs/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

var (
	store *RDB
)

func init() {
	conn, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	store = NewRDB(conn)
}

func TestRDB_Save(t *testing.T) {
	t.Run("Save url into PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawUrl, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		userId, err := store.SaveUser(context.Background(), user, rawUrl)
		require.NoError(t, err)
		require.NotEmpty(t, userId)
	})
	t.Run("Save url into PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawUrl, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		urlId, err := store.Save(context.Background(), rawUrl)
		require.NoError(t, err)
		require.NotEmpty(t, urlId)
	})
}

func TestRDB_Load(t *testing.T) {
	t.Run("Delete users from PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawUrl, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		urlId, err := store.SaveUser(context.Background(), user, rawUrl)
		require.NoError(t, err)
		urlFromDb, err := store.Load(context.Background(), urlId)
		require.NoError(t, err)
		require.NotEmpty(t, urlId)
		require.Equal(t, rawUrl, urlFromDb)
	})
}

func TestRDB_LoadUser(t *testing.T) {
	t.Run("Delete users from PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawUrl, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		require.NoError(t, err)
		urlId, err := store.SaveUser(context.Background(), user, rawUrl)
		require.NoError(t, err)
		urlFromDb, err := store.LoadUser(context.Background(), user, urlId)
		require.NoError(t, err)
		require.NotEmpty(t, urlId)
		require.Equal(t, rawUrl, urlFromDb)
	})
}

func TestRDB_DeleteUsers(t *testing.T) {
	t.Run("Delete users from PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawUrl, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		urlId, err := store.SaveUser(context.Background(), user, rawUrl)
		require.NoError(t, err)
		err = store.DeleteUsers(context.Background(), user, urlId)
		require.NoError(t, err)
		require.NotEmpty(t, urlId)
	})
}

func TestRDB_Ping(t *testing.T) {
	t.Run("Ping", func(t *testing.T) {
		err := store.Ping(context.Background())
		require.NoError(t, err)
	})
}

func TestRDB_Close(t *testing.T) {
	err := store.Close()
	require.NoError(t, err)
}
