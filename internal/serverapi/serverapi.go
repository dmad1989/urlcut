package serverapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dmad1989/urlcut/internal/logging"
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
	s.mux.Use(logging.WithLog)
	s.mux.Post("/", s.cutterHandler)
	s.mux.Get("/{path}", s.redirectHandler)
	s.mux.Post("/api/shorten", s.cutterJSONHandler)
}

func (s server) Run() error {
	logging.Log.Sugar().Infof("Server started at %s", s.config.GetURL())
	err := http.ListenAndServe(s.config.GetURL(), s.mux)
	if err != nil {
		return fmt.Errorf("serverapi.Run: %w", err)
	}
	return nil
}

type request struct {
	URL string `json:"url"`
}

type response struct {
	Result string `json:"result"`
}

func (s server) cutterJSONHandler(res http.ResponseWriter, req *http.Request) {
	var reqJSON request
	if req.Header.Get("Content-Type") != "application/json" {
		responseError(res, fmt.Errorf("cutterJsonHandler: content-type have to be application/json"))
		return
	}
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&reqJSON); err != nil {
		responseError(res, fmt.Errorf("cutterJsonHandler: decoding request: %w", err))
		return
	}
	code, err := s.cutterApp.Cut(reqJSON.URL)
	if err != nil {
		responseError(res, fmt.Errorf("cutterJsonHandler: getting code for url: %w", err))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	respJSON := response{
		Result: fmt.Sprintf("%s/%s", s.config.GetShortAddress(), code),
	}
	enc := json.NewEncoder(res)
	if err := enc.Encode(respJSON); err != nil {
		responseError(res, fmt.Errorf("cutterJsonHandler: encoding response: %w", err))
		return
	}
}

func (s server) cutterHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		responseError(res, fmt.Errorf("cutterHandler: reading request body: %w", err))
		return
	}

	if len(body) <= 0 {
		responseError(res, fmt.Errorf("cutterHandler: empty body not expected"))
		return
	}

	_, err = url.ParseRequestURI(string(body))
	if err != nil {
		responseError(res, fmt.Errorf("cutterHandler: parsing URI: %s : %w", string(body), err))
		return
	}

	code, err := s.cutterApp.Cut(string(body))
	if err != nil {
		responseError(res, fmt.Errorf("cutterHandler: getting code for url: %w", err))
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
		responseError(res, fmt.Errorf("redirectHandler: fetching url fo redirect: %w", err))
		return
	}
	http.Redirect(res, req, redirectURL, http.StatusTemporaryRedirect)
}

func responseError(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(err.Error()))
}
