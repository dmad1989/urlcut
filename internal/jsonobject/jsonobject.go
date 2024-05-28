// Package jsonobject contain dto for json.
// Objects processed to json using easyjson.
package jsonobject

//easyjson:json
type Item struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	ID          int    `json:"uuid"`
}

// Batch содержит список из URL
//
//easyjson:json
type Batch []BatchItem

//easyjson:json
type BatchItem struct {
	ID string `json:"correlation_id,omitempty" example:"1"`
	// URL для сокращения
	OriginalURL string `json:"original_url,omitempty" example:"http://ya.ru"`
	// Сокращенный URL
	ShortURL string `json:"short_url,omitempty" example:"http://localhost:8080/rjhsha"`
}

// Request содержит запрос с URL для сокращения
//
//easyjson:json
type Request struct {
	URL string `json:"url" example:"http://ya.ru"`
}

// Response содержит ответ с сокращенным URL
//
//easyjson:json
type Response struct {
	Result string `json:"result" example:"http://localhost:8080/rjhsha"`
}

// ShortIds содержит список из сокращений
//
//easyjson:json
type ShortIds []string

// Stats содержит данные статистики
//
//easyjson:json
type Stats struct {
	URLs  int `json:"urls" example:"1"`
	Users int `json:"users" example:"1"`
}
