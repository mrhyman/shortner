// Package api содержит объекты передачи данных, используемые в контрактах API.
package api

// ShortenRequest представляет тело запроса для сокращения одного URL.
// generate:reset
type ShortenRequest struct {
	// URL это оригинальный URL, который нужно сократить.
	URL string `json:"url"`
}

// ShortenBatchRequest представляет один элемент в пакетном запросе для сокращения нескольких URL.
type ShortenBatchRequest struct {
	// CorrelationID это уникальный идентификатор для отслеживания этого конкретного запроса в пакете.
	CorrelationID string `json:"correlation_id"`
	// OriginalURL это URL, который нужно сократить.
	OriginalURL string `json:"original_url"`
}

// ShortenResponse представляет тело ответа для операции сокращения URL.
type ShortenResponse struct {
	// Result содержит сокращенный URL.
	Result string `json:"result"`
}

// ShortenBatchResponse представляет один элемент в пакетном ответе для операций сокращения URL.
type ShortenBatchResponse struct {
	// CorrelationID это уникальный идентификатор, который связывает этот ответ с его запросом.
	CorrelationID string `json:"correlation_id"`
	// ShortURL это сокращенный URL, сгенерированный для оригинального URL.
	ShortURL string `json:"short_url"`
}

// UserLinksResponse представляет один элемент в ответе для получения ссылок пользователя.
type UserLinksResponse struct {
	// ShortURL это сокращенный URL.
	ShortURL string `json:"short_url"`
	// OriginalURL это оригинальный URL, который был сокращен.
	OriginalURL string `json:"original_url"`
}
