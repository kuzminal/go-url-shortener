package store

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"net/url"
	"strconv"
	"testing"
)

func BenchmarkInMemory_Save(b *testing.B) {
	ctx := context.Background()
	store := NewInMemory()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = store.Save(ctx, urlToStore)
	}
}

func BenchmarkInMemory_Load(b *testing.B) {
	ctx := context.Background()
	store := NewInMemory()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	id, _ := store.Save(ctx, urlToStore)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = store.Load(ctx, id)
	}
}

func BenchmarkInMemory_SaveBatch(b *testing.B) {
	ctx := context.Background()
	store := NewInMemory()
	urls := make([]*url.URL, 1000, 1000)
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	for i := 0; i < 1000; i++ {
		urls[i] = urlToStore
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = store.SaveBatch(ctx, urls)
	}
}

func BenchmarkInMemory_SaveUser(b *testing.B) {
	ctx := context.Background()
	store := NewInMemory()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	uuidToStore := uuid.Must(uuid.NewV4())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = store.SaveUser(ctx, uuidToStore, urlToStore)
	}
}

func BenchmarkInMemory_SaveUserBatch(b *testing.B) {
	ctx := context.Background()
	store := NewInMemory()
	urls := make([]*url.URL, 1000, 1000)
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	for i := 0; i < 1000; i++ {
		urls[i] = urlToStore
	}
	uuidToStore := uuid.Must(uuid.NewV4())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = store.SaveUserBatch(ctx, uuidToStore, urls)
	}
}

func BenchmarkInMemory_LoadUser(b *testing.B) {
	ctx := context.Background()
	store := NewInMemory()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	uuidToStore := uuid.Must(uuid.NewV4())
	id, _ := store.SaveUser(ctx, uuidToStore, urlToStore)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = store.LoadUser(ctx, uuidToStore, id)
	}
}

func BenchmarkInMemory_LoadUsers(b *testing.B) {
	ctx := context.Background()
	store := NewInMemory()
	urls := make([]*url.URL, 1000, 1000)
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	uuidToStore := uuid.Must(uuid.NewV4())
	for i := 0; i < 1000; i++ {
		urls[i] = urlToStore
	}
	_, _ = store.SaveUserBatch(ctx, uuidToStore, urls)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = store.LoadUsers(ctx, uuidToStore)
	}
}

func TestInMemory_Save(t *testing.T) {
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	m := NewInMemory()
	t.Run("save", func(t *testing.T) {
		gotId, err := m.Save(context.Background(), urlToStore)
		assert.NoError(t, err)
		assert.Equal(t, "0", gotId)
	})
}

func TestInMemory_Load(t *testing.T) {
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	m := NewInMemory()
	ctx := context.Background()
	gotId, _ := m.Save(ctx, urlToStore)

	t.Run("load regular", func(t *testing.T) {
		u, err := m.Load(ctx, gotId)
		assert.NoError(t, err)
		assert.Equal(t, u, urlToStore)
	})
	t.Run("load deleted", func(t *testing.T) {
		gotIdInt, _ := strconv.Atoi(gotId)
		m.store[gotIdInt] = nil
		_, err := m.Load(ctx, gotId)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDeleted)
	})
	t.Run("load not found", func(t *testing.T) {
		_, err := m.Load(ctx, "2")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestInMemory_SaveBatch(t *testing.T) {
	ctx := context.Background()
	store := NewInMemory()
	urls := make([]*url.URL, 1000, 1000)
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	for i := 0; i < 1000; i++ {
		urls[i] = urlToStore
	}
	t.Run("save regular", func(t *testing.T) {
		ids, err := store.SaveBatch(ctx, urls)
		assert.NoError(t, err)
		assert.Equal(t, len(ids), len(urls))
		assert.Contains(t, urls, urlToStore)
	})
}

func TestInMemory_SaveUser(t *testing.T) {
	ctx := context.Background()
	store := NewInMemory()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	uuidToStore := uuid.Must(uuid.NewV4())
	t.Run("save user", func(t *testing.T) {
		user, err := store.SaveUser(ctx, uuidToStore, urlToStore)
		assert.NoError(t, err)
		assert.Equal(t, "0", user)
	})
}

func TestInMemory_LoadUser(t *testing.T) {
	ctx := context.Background()
	store := NewInMemory()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	uuidToStore := uuid.Must(uuid.NewV4())
	id, _ := store.SaveUser(ctx, uuidToStore, urlToStore)
	t.Run("load user regular", func(t *testing.T) {
		urlStored, err := store.LoadUser(ctx, uuidToStore, id)
		assert.NoError(t, err)
		assert.Equal(t, urlStored, urlToStore)
	})
	t.Run("load user not found", func(t *testing.T) {
		_, err := store.LoadUser(ctx, uuidToStore, "2")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})
	t.Run("load user deleted", func(t *testing.T) {
		gotIdInt, _ := strconv.Atoi(id)
		urls := store.userStore[uuidToStore.String()]
		urls[gotIdInt] = nil
		_, err := store.LoadUser(ctx, uuidToStore, id)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDeleted)
	})
}

func TestInMemory_SaveUserBatch(t *testing.T) {
	ctx := context.Background()
	store := NewInMemory()
	urls := make([]*url.URL, 1000, 1000)
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	for i := 0; i < 1000; i++ {
		urls[i] = urlToStore
	}
	t.Run("save user batch regular", func(t *testing.T) {
		uuidToStore := uuid.Must(uuid.NewV4())
		batch, err := store.SaveUserBatch(ctx, uuidToStore, urls)
		assert.NoError(t, err)
		assert.Equal(t, len(batch), len(urls))
	})
}

func TestInMemory_LoadUsers(t *testing.T) {
	ctx := context.Background()
	store := NewInMemory()
	urls := make([]*url.URL, 1000, 1000)
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	uuidToStore := uuid.Must(uuid.NewV4())
	for i := 0; i < 1000; i++ {
		urls[i] = urlToStore
	}
	t.Run("load users regular", func(t *testing.T) {
		_, _ = store.SaveUserBatch(ctx, uuidToStore, urls)
		users, err := store.LoadUsers(ctx, uuidToStore)
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.Equal(t, len(urls), len(users))
	})
	t.Run("load user deleted", func(t *testing.T) {
		delete(store.userStore, uuidToStore.String())
		_, err := store.LoadUsers(ctx, uuidToStore)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestInMemory_DeleteUsers(t *testing.T) {
	ctx := context.Background()
	store := NewInMemory()
	urls := make([]*url.URL, 1000, 1000)
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	uuidToStore := uuid.Must(uuid.NewV4())
	ids := make([]string, 1000, 1000)
	for i := 0; i < 1000; i++ {
		urls[i] = urlToStore
		ids[i] = fmt.Sprintf("%d", i)
	}
	_, _ = store.SaveUserBatch(ctx, uuidToStore, urls)
	t.Run("load users regular", func(t *testing.T) {

		err := store.DeleteUsers(ctx, uuidToStore, ids...)
		assert.NoError(t, err)
	})
}

func TestNewInMemory(t *testing.T) {
	tests := []struct {
		name string
		want *InMemory
	}{
		{"reg", &InMemory{store: make([]*url.URL, 0, 10), userStore: make(map[string][]*url.URL)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewInMemory(), "NewInMemory()")
		})
	}
}
