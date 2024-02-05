package myjsons

//go:generate easyjson -all myjsons.go

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type StoreItem struct {
	ID          int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
