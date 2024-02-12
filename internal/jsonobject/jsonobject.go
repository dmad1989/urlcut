package jsonobject

//easyjson:json
type Item struct {
	ID          int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

//easyjson:json
type Batch []BatchItem

//easyjson:json
type BatchItem struct {
	ID          string `json:"correlation_id"`
	OriginalURL string `json:"original_url,omitempty"`
	ShortURL    string `json:"short_url,omitempty"`
}

//easyjson:json
type Request struct {
	URL string `json:"url"`
}

//easyjson:json
type Response struct {
	Result string `json:"result"`
}
