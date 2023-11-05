package store

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/gofrs/uuid"
)

var _ Store = (*InMemory)(nil)
var _ AuthStore = (*InMemory)(nil)

type InMemory struct {
	mu        sync.RWMutex
	store     map[string]*url.URL
	userStore map[string]map[string]*url.URL
}

// NewInMemory create new InMemory instance
func NewInMemory() *InMemory {
	return &InMemory{
		mu:        sync.RWMutex{},
		store:     make(map[string]*url.URL),
		userStore: make(map[string]map[string]*url.URL),
	}
}

func (m *InMemory) Save(_ context.Context, u *url.URL) (id string, err error) {
	m.mu.Lock()
	id = fmt.Sprintf("%x", len(m.store))
	m.store[id] = u
	m.mu.Unlock()
	return id, nil
}

func (m *InMemory) SaveBatch(_ context.Context, urls []*url.URL) (ids []string, err error) {
	for _, u := range urls {
		m.mu.Lock()
		id := fmt.Sprintf("%x", len(m.store))
		m.store[id] = u
		m.mu.Unlock()
		ids = append(ids, id)
	}
	if len(ids) != len(urls) {
		return nil, errors.New("not all URLs have been saved")
	}
	return
}

func (m *InMemory) Load(_ context.Context, id string) (u *url.URL, err error) {
	m.mu.RLock()
	u, ok := m.store[id]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	if u == nil {
		return nil, ErrDeleted
	}
	return u, nil
}

func (m *InMemory) SaveUser(ctx context.Context, uid uuid.UUID, u *url.URL) (id string, err error) {
	id, err = m.Save(ctx, u)
	if err != nil {
		return "", fmt.Errorf("cannot save URL to shared store: %w", err)
	}
	m.mu.Lock()
	if _, ok := m.userStore[uid.String()]; !ok {
		m.userStore[uid.String()] = make(map[string]*url.URL)
	}
	m.userStore[uid.String()][id] = u
	m.mu.Unlock()
	return id, nil
}

func (m *InMemory) SaveUserBatch(ctx context.Context, uid uuid.UUID, urls []*url.URL) (ids []string, err error) {
	ids, err = m.SaveBatch(ctx, urls)
	if err != nil {
		return nil, fmt.Errorf("cannot save URLs to shared store: %w", err)
	}
	m.mu.Lock()
	if _, ok := m.userStore[uid.String()]; !ok {
		m.userStore[uid.String()] = make(map[string]*url.URL)
	}
	for i, id := range ids {
		m.userStore[uid.String()][id] = urls[i]
	}
	m.mu.Unlock()
	return ids, nil
}

func (m *InMemory) LoadUser(ctx context.Context, uid uuid.UUID, id string) (u *url.URL, err error) {
	urls, err := m.LoadUsers(ctx, uid)
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

func (m *InMemory) LoadUsers(_ context.Context, uid uuid.UUID) (urls map[string]*url.URL, err error) {
	m.mu.RLock()
	urls, ok := m.userStore[uid.String()]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	// filter out deleted URLs
	res := make(map[string]*url.URL)
	for k, v := range urls {
		if v != nil {
			res[k] = v
		}
	}
	return res, nil
}

func (m *InMemory) DeleteUsers(_ context.Context, uid uuid.UUID, ids ...string) error {
	userID := uid.String()
	for _, id := range ids {
		m.mu.Lock()
		if _, ok := m.userStore[userID]; ok {
			m.store[id] = nil
			m.userStore[userID][id] = nil
		}
		m.mu.Unlock()
	}
	return nil
}

func (m *InMemory) Close() error {
	return nil
}

func (m *InMemory) Ping(_ context.Context) error {
	return nil
}
