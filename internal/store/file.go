package store

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/gofrs/uuid"
)

func init() {
	gob.Register(gobStore{})
}

var _ Store = (*FileStore)(nil)
var _ AuthStore = (*FileStore)(nil)

type gobStore struct {
	Hot     map[string]*url.URL
	UserHot map[string]map[string]*url.URL
}

// FileStore структура для файлового хранилища ссылок
type FileStore struct {
	mu      sync.RWMutex
	store   *gobStore
	enc     *gob.Encoder
	persist *os.File
}

// NewFileStore create new NewFileStore instance
func NewFileStore(filepath string) (*FileStore, error) {
	fd, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("cannot open file at path %s: %w", filepath, err)
	}

	gs := gobStore{
		Hot:     make(map[string]*url.URL),
		UserHot: make(map[string]map[string]*url.URL),
	}

	dec := gob.NewDecoder(fd)
	if err := dec.Decode(&gs); err != nil {
		// truncate bad file for reuse
		if err := fd.Truncate(0); err != nil {
			return nil, fmt.Errorf("cannot truncate broken storage file: %w", err)
		}
	}

	return &FileStore{
		store:   &gs,
		enc:     gob.NewEncoder(fd),
		persist: fd,
	}, nil
}

// Save сохраняем ссылку в файловом хранилище.
func (f *FileStore) Save(_ context.Context, u *url.URL) (id string, err error) {
	f.mu.Lock()
	id = fmt.Sprintf("%x", len(f.store.Hot))
	f.store.Hot[id] = u
	f.mu.Unlock()
	return id, f.flush()
}

// SaveBatch сохраняем ссылки в файловом хранилище.
func (f *FileStore) SaveBatch(_ context.Context, urls []*url.URL) (ids []string, err error) {
	f.mu.Lock()
	for _, u := range urls {
		id := fmt.Sprintf("%x", len(f.store.Hot))
		f.store.Hot[id] = u
		ids = append(ids, id)
	}
	f.mu.Unlock()
	if len(ids) != len(urls) {
		return nil, errors.New("not all URLs have been saved")
	}
	return ids, f.flush()
}

// Load загружаем ссылки из файлового хранилища по идентификатору.
func (f *FileStore) Load(_ context.Context, id string) (u *url.URL, err error) {
	f.mu.RLock()
	u, ok := f.store.Hot[id]
	f.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	if u == nil {
		return nil, ErrDeleted
	}
	return u, nil
}

// SaveUser сохраняем ссылку для пользователя
func (f *FileStore) SaveUser(ctx context.Context, uid uuid.UUID, u *url.URL) (id string, err error) {
	id, err = f.Save(ctx, u)
	if err != nil {
		return "", fmt.Errorf("cannot save URL to shared store: %w", err)
	}
	f.mu.Lock()
	if _, ok := f.store.UserHot[uid.String()]; !ok {
		f.store.UserHot[uid.String()] = make(map[string]*url.URL)
	}
	f.store.UserHot[uid.String()][id] = u
	f.mu.Unlock()
	return id, f.flush()
}

// SaveUserBatch сохраняем ссылки для пользователя
func (f *FileStore) SaveUserBatch(ctx context.Context, uid uuid.UUID, urls []*url.URL) (ids []string, err error) {
	ids, err = f.SaveBatch(ctx, urls)
	if err != nil {
		return nil, fmt.Errorf("cannot save URL to shared store: %w", err)
	}
	f.mu.Lock()
	if _, ok := f.store.UserHot[uid.String()]; !ok {
		f.store.UserHot[uid.String()] = make(map[string]*url.URL)
	}
	for i, id := range ids {
		f.store.UserHot[uid.String()][id] = urls[i]
	}
	f.mu.Unlock()
	return ids, f.flush()
}

// LoadUser загружаем ссылку для пользователя по ее идентификатору
func (f *FileStore) LoadUser(ctx context.Context, uid uuid.UUID, id string) (u *url.URL, err error) {
	urls, err := f.LoadUsers(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("cannot load user urls: %w", err)
	}
	u, ok := urls[id]
	if !ok {
		return nil, ErrNotFound
	}
	if u == nil {
		return nil, ErrDeleted
	}
	return u, nil
}

// LoadUsers загружаем ссылки для пользователя по их идентификаторам
func (f *FileStore) LoadUsers(_ context.Context, uid uuid.UUID) (urls map[string]*url.URL, err error) {
	f.mu.RLock()
	urls, ok := f.store.UserHot[uid.String()]
	f.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	res := make(map[string]*url.URL)
	for k, v := range urls {
		if v != nil {
			res[k] = v
		}
	}
	return res, nil
}

// DeleteUsers удаляем ссылки для пользователя.
func (f *FileStore) DeleteUsers(_ context.Context, uid uuid.UUID, ids ...string) error {
	userID := uid.String()
	f.mu.Lock()
	for _, id := range ids {
		if _, ok := f.store.UserHot[userID]; ok {
			f.store.Hot[id] = nil
			f.store.UserHot[userID][id] = nil
		}
	}
	f.mu.Unlock()
	return f.flush()
}

// Close закрываем файловое хранилище
func (f *FileStore) Close() error {
	if err := f.flush(); err != nil {
		return fmt.Errorf("cannot flush data to file: %w", err)
	}
	return f.persist.Close()
}

// Ping проверка работоспособности хранилища
func (f *FileStore) Ping(_ context.Context) error {
	if f.persist.Fd() == ^(uintptr(0)) {
		return errors.New("underlying file has been closed")
	}
	return nil
}

func (f *FileStore) flush() error {
	return f.enc.Encode(f.store)
}

// Users статистика по пользователям
func (f *FileStore) Users(_ context.Context) int {
	return len(f.store.UserHot)
}

// Urls статистика по ссылкам
func (f *FileStore) Urls(_ context.Context) int {
	return len(f.store.Hot)
}
