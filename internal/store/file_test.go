package store

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFileStore_Close(t *testing.T) {
	store, _ := NewFileStore("./store")
	t.Run("create file store", func(t *testing.T) {
		err := store.Close()
		assert.NoError(t, err)
	})
}

/*func TestFileStore_DeleteUsers(t *testing.T) {
	type fields struct {
		store   *gobStore
		enc     *gob.Encoder
		persist *os.File
	}
	type args struct {
		in0 context.Context
		uid uuid.UUID
		ids []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileStore{
				store:   tt.fields.store,
				enc:     tt.fields.enc,
				persist: tt.fields.persist,
			}
			tt.wantErr(t, f.DeleteUsers(tt.args.in0, tt.args.uid, tt.args.ids...), fmt.Sprintf("DeleteUsers(%v, %v, %v)", tt.args.in0, tt.args.uid, tt.args.ids...))
		})
	}
} */

func TestFileStore_Load(t *testing.T) {
	store, _ := NewFileStore("./store")
	ctx := context.Background()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/user")
	idFromStore, _ := store.Save(ctx, urlToStore)
	//uuidToStore, _ := uuid.NewV4()
	t.Run("save regular", func(t *testing.T) {
		urlFromStore, err := store.Load(ctx, idFromStore)
		assert.NoError(t, err)
		assert.Equal(t, urlToStore, urlFromStore)
	})
}

func TestFileStore_LoadUser(t *testing.T) {
	store, _ := NewFileStore("./store")
	ctx := context.Background()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/user")
	uuidToStore, _ := uuid.NewV4()
	idFromStore, _ := store.SaveUser(ctx, uuidToStore, urlToStore)
	t.Run("save regular", func(t *testing.T) {
		urlFromStore, err := store.LoadUser(ctx, uuidToStore, idFromStore)
		assert.NoError(t, err)
		assert.Equal(t, urlToStore, urlFromStore)
	})
}

func TestFileStore_LoadUsers(t *testing.T) {
	type fields struct {
		store   *gobStore
		enc     *gob.Encoder
		persist *os.File
	}
	type args struct {
		in0 context.Context
		uid uuid.UUID
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantUrls map[string]*url.URL
		wantErr  assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileStore{
				store:   tt.fields.store,
				enc:     tt.fields.enc,
				persist: tt.fields.persist,
			}
			gotUrls, err := f.LoadUsers(tt.args.in0, tt.args.uid)
			if !tt.wantErr(t, err, fmt.Sprintf("LoadUsers(%v, %v)", tt.args.in0, tt.args.uid)) {
				return
			}
			assert.Equalf(t, tt.wantUrls, gotUrls, "LoadUsers(%v, %v)", tt.args.in0, tt.args.uid)
		})
	}
}

/*
func TestFileStore_Ping(t *testing.T) {
	type fields struct {
		store   *gobStore
		enc     *gob.Encoder
		persist *os.File
	}
	type args struct {
		in0 context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileStore{
				store:   tt.fields.store,
				enc:     tt.fields.enc,
				persist: tt.fields.persist,
			}
			tt.wantErr(t, f.Ping(tt.args.in0), fmt.Sprintf("Ping(%v)", tt.args.in0))
		})
	}
}*/

func TestFileStore_Save(t *testing.T) {
	store, _ := NewFileStore("./store")
	ctx := context.Background()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	t.Run("save regular", func(t *testing.T) {
		gotID, err := store.Save(ctx, urlToStore)
		assert.NoError(t, err)
		assert.Equal(t, "0", gotID)
		assert.Equal(t, store.store.Hot["0"], urlToStore)
	})
}

func TestFileStore_SaveBatch(t *testing.T) {
	store, _ := NewFileStore("./store")
	ctx := context.Background()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/")
	urls := make([]*url.URL, 1000)
	for i := 0; i < 1000; i++ {
		urls[i] = urlToStore
	}
	t.Run("save batch regular", func(t *testing.T) {
		_, err := store.SaveBatch(ctx, urls)
		assert.NoError(t, err)
	})
}

func TestFileStore_SaveUser(t *testing.T) {
	store, _ := NewFileStore("./store")
	ctx := context.Background()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/user")
	uuidToStore, _ := uuid.NewV4()
	t.Run("save regular", func(t *testing.T) {
		gotID, err := store.SaveUser(ctx, uuidToStore, urlToStore)
		assert.NoError(t, err)
		assert.Equal(t, "0", gotID)
		assert.Equal(t, store.store.Hot["0"], urlToStore)
		assert.Equal(t, store.store.UserHot[uuidToStore.String()]["0"], urlToStore)
	})
}

func TestFileStore_SaveUserBatch(t *testing.T) {
	store, _ := NewFileStore("./store")
	ctx := context.Background()
	urlToStore, _ := url.Parse("https://practicum.yandex.ru/user")
	uuidToStore, _ := uuid.NewV4()
	urls := make([]*url.URL, 1000)
	for i := 0; i < 1000; i++ {
		urls[i] = urlToStore
	}
	t.Run("save regular", func(t *testing.T) {
		ids, err := store.SaveUserBatch(ctx, uuidToStore, urls)
		assert.NoError(t, err)
		assert.NotEmpty(t, ids)
	})
}

/*
func TestNewFileStore(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    *FileStore
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFileStore(tt.args.filepath)
			if !tt.wantErr(t, err, fmt.Sprintf("NewFileStore(%v)", tt.args.filepath)) {
				return
			}
			assert.Equalf(t, tt.want, got, "NewFileStore(%v)", tt.args.filepath)
		})
	}
}
}*/
