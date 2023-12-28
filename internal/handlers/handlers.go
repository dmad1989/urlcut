package handlers

import (
	"fmt"
	"io"
	"net/http"

	urls "github.com/dmad1989/urlcut/internal/urls"
)

func Manage(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		cutter(res, req)
	case http.MethodGet:
		redirect(res, req)
	default:
		res.WriteHeader(http.StatusBadRequest)
	}
}

func cutter(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(err.Error()))
		return
	}
	code := urls.Cut(body)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(fmt.Sprintf("http://%s%s%s", req.Host, req.URL.Path, code)))
}

func redirect(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	path := req.URL.Path[1:]
	redirectUrl := urls.GetUrls().GetKey(path)
	if redirectUrl == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, redirectUrl, http.StatusTemporaryRedirect)
}
