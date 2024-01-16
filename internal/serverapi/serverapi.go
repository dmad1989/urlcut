package serverapi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/dmad1989/urlcut/internal/config"
	"github.com/go-chi/chi/v5"
)

type worker interface {
	Cut(body []byte) (generated string, err error)
	GetKeyByValue(value string) string
}

type server struct {
	cut worker
	mux *chi.Mux
}

func New(cut worker) *server {
	api := &server{cut: cut, mux: chi.NewMux()}
	api.initHandlers()
	return api
}

func (api server) initHandlers() {
	api.mux.Post("/", api.cutterHandler)
	api.mux.Get("/{code}", api.redirectHandler)
	api.mux.MethodNotAllowed(api.errorHandler)
	api.mux.NotFound(api.errorHandler)
}

func (api server) Run() {
	fmt.Println(config.Conf.URL)
	err := http.ListenAndServe(config.Conf.URL, api.mux)
	if err != nil {
		panic(err)
	}
}

func (api server) errorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("wrong http method"))
}

func (api server) cutterHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		responseError(res, fmt.Errorf("wrong http method"))
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		responseError(res, err)
		return
	}

	if len(body) <= 0 {
		responseError(res, fmt.Errorf("empty body not expected"))
		return
	}

	_, err = url.ParseRequestURI(string(body))
	if err != nil {
		responseError(res, err)
		return
	}

	code, err := api.cut.Cut(body)
	if err != nil {
		responseError(res, err)
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	url := config.Conf.ShortAddress
	pattern := "%s/%s"
	if !strings.Contains(url, "http://") && !strings.Contains(url, "http//") {
		pattern = "http://%s/%s"
	}
	res.Write([]byte(fmt.Sprintf(pattern, url, code)))
}

func (api server) redirectHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		responseError(res, fmt.Errorf("wrong http method"))
		return
	}
	path := chi.URLParam(req, "code")
	if len(path) == 0 {
		responseError(res, fmt.Errorf("url path is empty"))
		return
	}

	redirectURL := api.cut.GetKeyByValue(path)
	if redirectURL == "" {
		responseError(res, fmt.Errorf("requested url not found"))
		return
	}
	http.Redirect(res, req, redirectURL, http.StatusTemporaryRedirect)
}

func responseError(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(err.Error()))
}
