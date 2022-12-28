package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"net/url"
	"strconv"
)

var _ Store = (*InMemory)(nil)
var _ AuthStore = (*InMemory)(nil)

type InMemory struct {
	store     []*url.URL
	userStore map[string][]*url.URL
}

// NewInMemory create new InMemory instance
func NewInMemory() *InMemory {
	return &InMemory{
		store:     make([]*url.URL, 0, 10),
		userStore: make(map[string][]*url.URL),
	}
}

func (m *InMemory) Save(_ context.Context, u *url.URL) (id string, err error) {
	id = strconv.Itoa(len(m.store))
	m.store = append(m.store, u)
	return id, nil
}

func (m *InMemory) SaveBatch(_ context.Context, urls []*url.URL) (ids []string, err error) {
	ids = make([]string, len(urls), len(urls))
	for i, u := range urls {
		id := strconv.Itoa(len(m.store))
		m.store = append(m.store, u)
		ids[i] = id
	}
	if len(ids) != len(urls) {
		return nil, errors.New("not all URLs have been saved")
	}
	return
}

func (m *InMemory) Load(_ context.Context, id string) (u *url.URL, err error) {
	indId, err := strconv.Atoi(id)
	if err != nil || indId > len(m.store) || indId < 0 {
		return nil, ErrNotFound
	}
	u = m.store[indId]
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
	if _, ok := m.userStore[uid.String()]; !ok {
		m.userStore[uid.String()] = make([]*url.URL, 0, 10)
	}
	m.userStore[uid.String()] = append(m.userStore[uid.String()], u)
	//m.userStore[uid.String()][idInt] = u
	return id, nil
}

func (m *InMemory) SaveUserBatch(ctx context.Context, uid uuid.UUID, urls []*url.URL) (ids []string, err error) {
	ids, err = m.SaveBatch(ctx, urls)
	if err != nil {
		return nil, fmt.Errorf("cannot save URLs to shared store: %w", err)
	}
	if _, ok := m.userStore[uid.String()]; !ok {
		m.userStore[uid.String()] = make([]*url.URL, 0, 10)
	}
	for i := range ids {
		m.userStore[uid.String()] = append(m.userStore[uid.String()], urls[i])
	}
	return ids, nil
}

func (m *InMemory) LoadUser(ctx context.Context, uid uuid.UUID, id string) (u *url.URL, err error) {
	urls, err := m.LoadUsers(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("cannot load user urls: %w", err)
	}

	indId, err := strconv.Atoi(id)
	if err != nil || indId > len(m.store) || indId < 0 {
		return nil, ErrNotFound
	}
	u = urls[id]
	if u == nil {
		return nil, ErrDeleted
	}

	return u, nil
}

func (m *InMemory) LoadUsers(_ context.Context, uid uuid.UUID) (map[string]*url.URL, error) {
	urls, ok := m.userStore[uid.String()]
	if !ok {
		return nil, ErrNotFound
	}
	// filter out deleted URLs
	res := make(map[string]*url.URL)
	for k, v := range urls {
		if v != nil {
			res[strconv.Itoa(k)] = v
		}
	}
	return res, nil
}

func (m *InMemory) DeleteUsers(_ context.Context, uid uuid.UUID, ids ...string) error {
	for _, id := range ids {
		userID := uid.String()
		if _, ok := m.userStore[userID]; ok {
			idInt, _ := strconv.Atoi(id)
			m.store[idInt] = nil
			m.userStore[userID][idInt] = nil
		}
	}
	return nil
}

func (m *InMemory) Close() error {
	return nil
}

func (m *InMemory) Ping(_ context.Context) error {
	return nil
}
