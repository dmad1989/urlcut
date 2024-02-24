package serverapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/dmad1989/urlcut/internal/config"
	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/dmad1989/urlcut/internal/logging"
	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"
)

type app interface {
	Cut(cxt context.Context, url string) (generated string, err error)
	GetKeyByValue(cxt context.Context, value string) (res string, err error)
	PingDB(context.Context) error
	UploadBatch(ctx context.Context, batch jsonobject.Batch) (jsonobject.Batch, error)
	GetUserURLs(ctx context.Context) (jsonobject.Batch, error)
}

type conf interface {
	GetURL() string
	GetShortAddress() string
	GetUserContextKey() config.ContextKey
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
	s.mux.Use(logging.WithLog, s.Auth, gzipMiddleware)
	s.mux.Post("/", s.cutterHandler)
	s.mux.Get("/{path}", s.redirectHandler)
	s.mux.Get("/ping", s.pingHandler)
	s.mux.Post("/api/shorten", s.cutterJSONHandler)
	s.mux.Post("/api/shorten/batch", s.cutterJSONBatchHandler)
	s.mux.Get("/api/user/urls", s.userUrlsHandler)
}

func (s server) Run(ctx context.Context) error {
	defer logging.Log.Sync()
	logging.Log.Infof("Server started at %s", s.config.GetURL())
	httpServer := &http.Server{
		Addr:    s.config.GetURL(),
		Handler: s.mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		err := httpServer.ListenAndServe()
		if err != nil {
			return fmt.Errorf("serverapi.Run: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		<-gCtx.Done()
		return httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}
	return nil
}

func (s server) cutterJSONHandler(res http.ResponseWriter, req *http.Request) {
	var reqJSON jsonobject.Request
	if req.Header.Get("Content-Type") != "application/json" {
		responseError(res, fmt.Errorf("cutterJsonHandler: content-type have to be application/json"))
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		responseError(res, fmt.Errorf("cutterJsonHandler: reading request body: %w", err))
		return
	}
	if err := reqJSON.UnmarshalJSON(body); err != nil {
		responseError(res, fmt.Errorf("cutterJsonHandler: decoding request: %w", err))
		return
	}
	code, err := s.cutterApp.Cut(req.Context(), reqJSON.URL)
	status := http.StatusCreated
	if err != nil {
		var uerr *cutter.UniqueURLError
		if !errors.As(err, &uerr) {
			responseError(res, fmt.Errorf("cutterJsonHandler: getting code for url: %w", err))
			return
		}
		status = http.StatusConflict
		code = uerr.Code
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	respJSON := jsonobject.Response{
		Result: fmt.Sprintf("%s/%s", s.config.GetShortAddress(), code),
	}
	respb, err := respJSON.MarshalJSON()
	if err != nil {
		responseError(res, fmt.Errorf("cutterJsonHandler: encoding response: %w", err))
		return
	}
	res.Write(respb)
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

	code, err := s.cutterApp.Cut(req.Context(), string(body))
	status := http.StatusCreated
	if err != nil {
		var uerr *cutter.UniqueURLError
		if !errors.As(err, &uerr) {
			responseError(res, fmt.Errorf("cutterHandler: getting code for url: %w", err))
			return
		}
		status = http.StatusConflict
		code = uerr.Code
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(status)
	res.Write([]byte(fmt.Sprintf("%s/%s", s.config.GetShortAddress(), code)))
}

func (s server) redirectHandler(res http.ResponseWriter, req *http.Request) {
	path := chi.URLParam(req, "path")
	if path == "" {
		responseError(res, fmt.Errorf("redirectHandler: url path is empty"))
		return
	}

	redirectURL, err := s.cutterApp.GetKeyByValue(req.Context(), path)
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

func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextW := w
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				responseError(w, fmt.Errorf("gzip: read compressed body: %w", err))
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header()
			cw := newCompressWriter(w)
			nextW = cw
			defer cw.Close()
		}

		h.ServeHTTP(nextW, r)
	})
}

func (s server) pingHandler(res http.ResponseWriter, req *http.Request) {
	err := s.cutterApp.PingDB(req.Context())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(fmt.Errorf("pingHandler : %w", err).Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (s server) cutterJSONBatchHandler(res http.ResponseWriter, req *http.Request) {
	var batchRequest jsonobject.Batch
	if req.Header.Get("Content-Type") != "application/json" {
		responseError(res, fmt.Errorf("JSONBatchHandler: content-type have to be application/json"))
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		responseError(res, fmt.Errorf("JSONBatchHandler: reading request body: %w", err))
		return
	}

	if err := batchRequest.UnmarshalJSON(body); err != nil {
		responseError(res, fmt.Errorf("JSONBatchHandler: decoding request: %w", err))
		return
	}
	logging.Log.Info(batchRequest)
	batchResponse, err := s.cutterApp.UploadBatch(req.Context(), batchRequest)

	if err != nil {
		responseError(res, fmt.Errorf("JSONBatchHandler: getting code for url: %w", err))
		return
	}

	for i := 0; i < len(batchResponse); i++ {
		batchResponse[i].ShortURL = fmt.Sprintf("%s/%s", s.config.GetShortAddress(), batchResponse[i].ShortURL)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	respb, err := batchResponse.MarshalJSON()
	if err != nil {
		responseError(res, fmt.Errorf("JSONBatchHandler: encoding response: %w", err))
		return
	}
	res.Write(respb)
}

func (s server) userUrlsHandler(res http.ResponseWriter, req *http.Request) {
	urls, err := s.cutterApp.GetUserURLs(req.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			res.WriteHeader(http.StatusNoContent)
			return
		}
		responseError(res, fmt.Errorf("userUrlsHandler: getting all urls: %w", err))
	}

	if len(urls) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	for i := 0; i < len(urls); i++ {
		urls[i].ShortURL = fmt.Sprintf("%s/%s", s.config.GetShortAddress(), urls[i].ShortURL)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	respb, err := urls.MarshalJSON()
	if err != nil {
		responseError(res, fmt.Errorf("userUrlsHandler: encoding response: %w", err))
		return
	}
	res.Write(respb)
}
