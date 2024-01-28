package myjsons

//go:generate easyjson -all myjsons.go
type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}
