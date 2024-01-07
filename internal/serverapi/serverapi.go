package serverapi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

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
	router := api.mux
	router.Post("/", api.cutterHandler)
	router.Get("/{code}", api.redirectHandler)
	fmt.Println("start init router ", router)
	router.MethodNotAllowed(api.errorHandler)
	router.NotFound(api.errorHandler)
}

func (api server) Run() {
	err := http.ListenAndServe(`:8080`, api.mux)
	fmt.Println("main err:", err)
	if err != nil {
		panic(err)
	}
}

func (api server) errorHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("inside of func ")
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
	res.Write([]byte(fmt.Sprintf("http://%s%s%s", req.Host, req.URL.Path, code)))
}

func (api server) redirectHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		responseError(res, fmt.Errorf("wrong http method"))
		return
	}
	fmt.Println("URL.Path:", len(req.URL.Path[1:]), ";")
	path := chi.URLParam(req, "code")
	fmt.Println("chi.URLParam ", len(path), ";")
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
