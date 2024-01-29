package myjsons

//easyjson:json
type Request struct {
	URL string `json:"url"`
}

//easyjson:json
type Response struct {
	Result string `json:"result"`
}

//easyjson:json
type StoreItem struct {
	ID          int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

//easyjson:json
type StoreItemSlice []StoreItem
