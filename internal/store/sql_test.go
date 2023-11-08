package store

import (
	"context"
	"database/sql"
	"github.com/gofrs/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	store *RDB
)

func TestMain(m *testing.M) {
	container, err := CreatePostgresContainer(context.Background())
	if err != nil {
		panic(err)
	}
	connStr, err := container.ConnectionString(context.Background(), "sslmode=disable")
	if err != nil {
		panic(err)
	}
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	store = NewRDB(conn)
	store.Bootstrap(context.Background())
	code := m.Run()
	container.Terminate(context.Background())
	os.Exit(code)
}

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

func TestRDB_Save(t *testing.T) {
	t.Run("Save url into PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawURL, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		userID, err := store.SaveUser(context.Background(), user, rawURL)
		require.NoError(t, err)
		require.NotEmpty(t, userID)
	})
	t.Run("Save url into PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawURL, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		urlID, err := store.Save(context.Background(), rawURL)
		require.NoError(t, err)
		require.NotEmpty(t, urlID)
	})
}

func TestRDB_Load(t *testing.T) {
	t.Run("Delete users from PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawURL, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		urlID, err := store.SaveUser(context.Background(), user, rawURL)
		require.NoError(t, err)
		urlFromDB, err := store.Load(context.Background(), urlID)
		require.NoError(t, err)
		require.NotEmpty(t, urlID)
		require.Equal(t, rawURL, urlFromDB)
	})
}

func TestRDB_LoadUser(t *testing.T) {
	t.Run("Delete users from PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawURL, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		require.NoError(t, err)
		urlID, err := store.SaveUser(context.Background(), user, rawURL)
		require.NoError(t, err)
		urlFromDB, err := store.LoadUser(context.Background(), user, urlID)
		require.NoError(t, err)
		require.NotEmpty(t, urlID)
		require.Equal(t, rawURL, urlFromDB)
	})
}

func TestRDB_DeleteUsers(t *testing.T) {
	t.Run("Delete users from PG store", func(t *testing.T) {
		user, err := uuid.NewV4()
		require.NoError(t, err)
		rawURL, _ := url.Parse("https://practicum.yandex.ru/" + user.String())
		urlID, err := store.SaveUser(context.Background(), user, rawURL)
		require.NoError(t, err)
		err = store.DeleteUsers(context.Background(), user, urlID)
		require.NoError(t, err)
		require.NotEmpty(t, urlID)
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
