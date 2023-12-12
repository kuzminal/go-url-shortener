package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/config"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
	"net"
	"net/url"

	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/auth"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
)

// Shorten обработчик сокращения ссылки
func (i *Instance) Shorten(ctx context.Context, rawURL string) (shortURL string, err error) {
	uid := auth.UIDFromContext(ctx)

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrParseURL
	}
	var id string
	if uid != nil {
		id, err = i.Store.SaveUser(ctx, *uid, u)
	} else {
		id, err = i.Store.Save(ctx, u)
	}

	if err != nil && !errors.Is(err, store.ErrConflict) {
		return "", fmt.Errorf("cannot save URL to storage: %w", err)
	}
	return fmt.Sprintf("%s/%s", i.BaseURL, id), err
}

// ShortenBatch пакетный обработчик сокращения ссылок
func (i *Instance) ShortenBatch(ctx context.Context, rawURLs []*url.URL) (shortURLs []string, err error) {
	uid := auth.UIDFromContext(ctx)

	var ids []string
	if uid != nil {
		ids, err = i.Store.SaveUserBatch(ctx, *uid, rawURLs)
	} else {
		ids, err = i.Store.SaveBatch(ctx, rawURLs)
	}

	if err != nil {
		return nil, fmt.Errorf("cannot save URL to storage: %w", err)
	}

	for _, id := range ids {
		shortURLs = append(shortURLs, fmt.Sprintf("%s/%s", i.BaseURL, id))
	}

	return shortURLs, nil
}

// Statistics предсоатвляет статистику по ссылкам и пользователям
func (i *Instance) Statistics(ctx context.Context, ip string) (models.Statistics, error) {
	_, ipNet, _ := net.ParseCIDR(config.TrustedSubnet)
	if ipNet == nil || !ipNet.Contains(net.ParseIP(ip)) {
		return models.Statistics{}, fmt.Errorf("forbidden for ip %s", ip)
	}
	var res models.Statistics
	statUsers := i.Store.Users(ctx)
	statUrls := i.Store.Urls(ctx)
	res.Users = statUsers
	res.Urls = statUrls

	return res, nil
}

// LoadURL поиск ссылки в хранилище
func (i *Instance) LoadURL(ctx context.Context, id string) (*url.URL, error) {
	u, err := i.Store.Load(ctx, id)
	if err != nil {
		return &url.URL{}, err
	}
	return u, nil
}

// LoadUsers загрузка ссылок пользователя
func (i *Instance) LoadUsers(ctx context.Context) ([]models.URLResponse, error) {
	uid := auth.UIDFromContext(ctx)
	if uid == nil {
		return []models.URLResponse{}, ErrAuth
	}
	urls, err := i.Store.LoadUsers(ctx, *uid)
	if err != nil {
		return []models.URLResponse{}, err
	}
	var resp []models.URLResponse
	for id, u := range urls {
		resp = append(resp, models.URLResponse{
			ShortURL:    i.BaseURL + "/" + id,
			OriginalURL: u.String(),
		})
	}
	return resp, nil
}

// Ping проверка работоспособности приложения
func (i *Instance) Ping(ctx context.Context) error {
	for j := 0; j < 3; j++ {
		if err := i.Store.Ping(ctx); err != nil {
			return err
		}
	}
	return nil
}

// BatchShorten пакетное соркащение ссылок
func (i *Instance) BatchShorten(req []models.BatchShortenRequest, ctx context.Context) ([]models.BatchShortenResponse, error) {
	var urls []*url.URL
	for _, pair := range req {
		u, errs := url.Parse(pair.OriginalURL)
		if errs != nil {
			return []models.BatchShortenResponse{}, ErrParseURL
		}
		urls = append(urls, u)
	}

	shortURLs, err := i.ShortenBatch(ctx, urls)
	if err != nil {
		return []models.BatchShortenResponse{}, err
	}

	if len(shortURLs) != len(req) {
		return []models.BatchShortenResponse{}, ErrURLLength
	}

	res := make([]models.BatchShortenResponse, 0, len(shortURLs))
	for i, shortURL := range shortURLs {
		res = append(res, models.BatchShortenResponse{
			CorrelationID: req[i].CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return res, nil
}
