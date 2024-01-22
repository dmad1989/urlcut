package serverapi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

type app interface {
	Cut(url string) (generated string, err error)
	GetKeyByValue(value string) (res string, err error)
}

type conf interface {
	GetURL() string
	GetShortAddress() string
}

type server struct {
	cutterApp app
	config    conf
	mux       *chi.Mux
}

func New(cutApp app, config conf) *server {
	api := &server{cutterApp: cutApp, config: config, mux: chi.NewMux()}
	api.initHandlers()
	return api
}

func (s server) initHandlers() {
	s.mux.Post("/", s.cutterHandler)
	s.mux.Get("/{path}", s.redirectHandler)
}

func (s server) Run() {
	err := http.ListenAndServe(s.config.GetURL(), s.mux)
	if err != nil {
		panic(err)
	}
}

func (s server) cutterHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		responseError(res, fmt.Errorf("cutterHandler: error while read request body: %w", err))
		return
	}

	if len(body) <= 0 {
		responseError(res, fmt.Errorf("cutterHandler: empty body not expected"))
		return
	}

	_, err = url.ParseRequestURI(string(body))
	if err != nil {
		responseError(res, fmt.Errorf("cutterHandler: error while parsing URI: %s : %w", string(body), err))
		return
	}

	code, err := s.cutterApp.Cut(string(body))
	if err != nil {
		responseError(res, fmt.Errorf("cutterHandler: error while getting code for url: %w", err))
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(fmt.Sprintf("%s/%s", s.config.GetShortAddress(), code)))
}

func (s server) redirectHandler(res http.ResponseWriter, req *http.Request) {
	path := chi.URLParam(req, "path")
	if path == "" {
		responseError(res, fmt.Errorf("redirectHandler: url path is empty"))
		return
	}

	redirectURL, err := s.cutterApp.GetKeyByValue(path)
	if err != nil {
		responseError(res, fmt.Errorf("redirectHandler: error while fetching url fo redirect: %w", err))
		return
	}
	http.Redirect(res, req, redirectURL, http.StatusTemporaryRedirect)
}

func responseError(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(err.Error()))
}
