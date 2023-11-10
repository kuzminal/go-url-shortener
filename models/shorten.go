// Package models содержит структуры, которые используются для передачи в API.
package models

// ShortenRequest запрос на сокращение ссылки.
//
// Используется в запросе, например, так:
//
//	 func (i *Instance) ShortenAPIHandler(w http.ResponseWriter, r *http.Request) {
//		var req models.ShortenRequest
//		err := json.NewDecoder(r.Body).Decode(&req)
//		...
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse ответ с сокращенной ссылкой.
type ShortenResponse struct {
	Result string `json:"result"`
}

// URLResponse ответ с сокращенной и оригинальной ссылками.
type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// BatchShortenRequest запрос на сокращение нескольких ссылок.
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchShortenResponse ответ на запрос на сокращение нескольких ссылок.
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
