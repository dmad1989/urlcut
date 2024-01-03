package serverApi

import (
	"fmt"
	"io"
	"net/http"
)

type cutter interface {
	Cut(body []byte) (generated string)
	GetKeyByValue(value string) string
}

type server struct {
	cut cutter
	mux *http.ServeMux
}

func New(cut cutter) *server {
	api := &server{cut: cut, mux: http.NewServeMux()}
	api.mux.HandleFunc(`/`, api.initHandlers)
	return api
}

func (api server) initHandlers(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		api.cutterHandler(res, req)
	case http.MethodGet:
		api.redirectHandler(res, req)
	default:
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (api server) Run() {
	err := http.ListenAndServe(`:8080`, api.mux)
	fmt.Println("main err:", err)
	if err != nil {
		panic(err)
	}
}

func (api server) cutterHandler(res http.ResponseWriter, req *http.Request) {
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
	code := api.cut.Cut(body)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(fmt.Sprintf("http://%s%s%s", req.Host, req.URL.Path, code)))
}

func (api server) redirectHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	path := req.URL.Path[1:]
	redirectURL := api.cut.GetKeyByValue(path)
	if redirectURL == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, redirectURL, http.StatusTemporaryRedirect)
}
