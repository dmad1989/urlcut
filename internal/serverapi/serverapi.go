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

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"

	"github.com/dmad1989/urlcut/internal/config"
	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/dbstore"
	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/dmad1989/urlcut/internal/logging"
)

// @Title URLCutter API
// @Description Сервис сокращения ссылок.
// @Version 1.0

// @Contact.email dmad1989@gmail.com

// @SecurityDefinitions.token tokenAuth
// @In cookie
// @Name token

// @Tag.name Cut
// @Tag.description "Группа запросов для сокращения URL"

// @Tag.name UserURLs
// @Tag.description "Группа запросов для работы с URL пользователя"

// @Tag.name Operate
// @Tag.description "Группа запросов для работы с сокращенными URL"

// @Tag.name Info
// @Tag.description "Группа запросов состояния сервиса"

// ICutter интерфейс слоя с бизнес логикой
type ICutter interface {
	Cut(cxt context.Context, url string) (generated string, err error)
	GetKeyByValue(cxt context.Context, value string) (res string, err error)
	PingDB(context.Context) error
	UploadBatch(ctx context.Context, batch jsonobject.Batch) (jsonobject.Batch, error)
	GetUserURLs(ctx context.Context) (jsonobject.Batch, error)
	DeleteUrls(userID string, ids jsonobject.ShortIds)
}

// Configer интерйфейс конфигураци
type Configer interface {
	GetURL() string
	GetShortAddress() string
}

// Server содержит интерфейсы для обращения к другим слоям и роутинг.
type Server struct {
	cutter ICutter
	config Configer
	mux    *chi.Mux
}

// New создает новый Server и инициализирует Хэндлеры.
func New(cutter ICutter, config Configer) *Server {
	api := &Server{cutter: cutter, config: config, mux: chi.NewMux()}
	api.initHandlers()
	return api
}

// Run запускает сервер в отдельной горутине.
// В другой горутине ожидает сигнала от контекста о завершении, чтобы отключить сервер.
// Пишет ошибку в консоль, о причине выключения.
func (s Server) Run(ctx context.Context) error {
	defer func() {
		err := logging.Log.Sync()
		if err != nil {
			logging.Log.Fatalf("log.sync in run: %w", err)
		}
	}()
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

func (s Server) initHandlers() {
	s.mux.Use(logging.WithLog, s.Auth, gzipMiddleware)
	s.mux.Mount("/debug", middleware.Profiler())
	s.mux.Post("/", s.cutterHandler)
	s.mux.Get("/{path}", s.redirectHandler)
	s.mux.Get("/ping", s.pingHandler)
	s.mux.Post("/api/shorten", s.cutterJSONHandler)
	s.mux.Post("/api/shorten/batch", s.cutterJSONBatchHandler)
	s.mux.Get("/api/user/urls", s.userUrlsHandler)
	s.mux.Delete("/api/user/urls", s.deleteUserUrlsHandler)
}

// cutterJSONHandler godoc
// @Tags Cut
// @Summary Запрос на сокращение URL
// @ID cutterJSON
// @Accept  json
// @Produce json
// @Success 201 {object} jsonobject.Response
// @Failure 401 {string} string "Ошибка авторизации"
// @Failure 400 {string} string "Ошибка"
// @Router /api/shorten [post]
func (s Server) cutterJSONHandler(res http.ResponseWriter, req *http.Request) {
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
	err = reqJSON.UnmarshalJSON(body)
	if err != nil {
		responseError(res, fmt.Errorf("cutterJsonHandler: decoding request: %w", err))
		return
	}
	code, err := s.cutter.Cut(req.Context(), reqJSON.URL)
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

// cutterHandler godoc
// @Tags Cut
// @Summary Запрос на сокращение URL
// @ID cutterText
// @Accept  plain/text
// @Produce plain/text
// @Success 201 {string} string "Сокращенный URL"
// @Failure 401 {string} string "Ошибка авторизации"
// @Failure 400 {string} string "Ошибка"
// @Router / [post]
func (s Server) cutterHandler(res http.ResponseWriter, req *http.Request) {
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

	code, err := s.cutter.Cut(req.Context(), string(body))
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

// redirectHandler godoc
// @Tags Operate
// @Summary Переход по сокращеному URL
// @ID redirect
// @Accept  plain/text
// @Param path path string true "Сокращенный url"
// @Success 307 "Переход по сокращенному URL"
// @Failure 401 {string} string "Ошибка авторизации"
// @Failure 410 {string} string "url was deleted"
// @Failure 400 {string} string "Ошибка"
// @Router /{path} [get]
func (s Server) redirectHandler(res http.ResponseWriter, req *http.Request) {
	path := chi.URLParam(req, "path")
	if path == "" {
		responseError(res, fmt.Errorf("redirectHandler: url path is empty"))
		return
	}

	redirectURL, err := s.cutter.GetKeyByValue(req.Context(), path)
	if err != nil {
		if errors.Is(err, dbstore.ErrorDeletedURL) {
			res.WriteHeader(http.StatusGone)
			res.Write([]byte(err.Error()))
			return
		}
		responseError(res, fmt.Errorf("redirectHandler: fetching url fo redirect: %w", err))
		return
	}
	http.Redirect(res, req, redirectURL, http.StatusTemporaryRedirect)
}

// pingHandler godoc
// @Tags Info
// @Summary Проверка соединения с БД
// @ID ping
// @Accept  */*
// @Produce plain/text
// @Success 200 {string} string
// @Failure 401 {string} string "Ошибка авторизации"
// @Failure 500 {string} string "Ошибка"
// @Router /ping [get]
func (s Server) pingHandler(res http.ResponseWriter, req *http.Request) {
	err := s.cutter.PingDB(req.Context())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(fmt.Errorf("pingHandler : %w", err).Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
}

// cutterJSONBatchHandler godoc
// @Tags Cut
// @Summary Запрос на сокращение списка URL
// @ID cutterBatch
// @Accept  json
// @Produce json
// @Success 201 {object} jsonobject.Batch
// @Failure 401 {string} string "Ошибка авторизации"
// @Failure 400 {string} string "Ошибка"
// @Router /api/shorten/batch [post]
func (s Server) cutterJSONBatchHandler(res http.ResponseWriter, req *http.Request) {
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

	err = batchRequest.UnmarshalJSON(body)
	if err != nil {
		responseError(res, fmt.Errorf("JSONBatchHandler: decoding request: %w", err))
		return
	}
	logging.Log.Info(batchRequest)
	batchResponse, err := s.cutter.UploadBatch(req.Context(), batchRequest)

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

// userUrlsHandler godoc
// @Tags UserURLs
// @Summary Все скоращенные URL текущего пользователя
// @ID userURLs
// @Produce plain/text
// @Success 200 {object} jsonobject.Batch
// @Failure 401 {string} string "Ошибка авторизации"
// @Failure 204 {string} string "Нет сокращенных URL"
// @Failure 400 {string} string "Ошибка"
// @Router /api/user/urls [get]
func (s Server) userUrlsHandler(res http.ResponseWriter, req *http.Request) {
	err, _ := req.Context().Value(config.ErrorCtxKey).(error)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	urls, err := s.cutter.GetUserURLs(req.Context())
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

// deleteUserUrlsHandler godoc
// @Tags UserURLs
// @Summary Запрос на удаление сокращеных URL
// @ID deleteUserUrls
// @Accept  json
// @Produce plain/text
// @Success 202 {object} jsonobject.ShortIds
// @Failure 400 {string} string "Ошибка"
// @Router / [post]
func (s Server) deleteUserUrlsHandler(res http.ResponseWriter, req *http.Request) {
	err, _ := req.Context().Value(config.ErrorCtxKey).(error)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}
	var ids jsonobject.ShortIds
	if req.Header.Get("Content-Type") != "application/json" {
		responseError(res, fmt.Errorf("deleteUserUrlsHandler: content-type have to be application/json"))
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		responseError(res, fmt.Errorf("deleteUserUrlsHandler: reading request body: %w", err))
		return
	}

	if err := ids.UnmarshalJSON(body); err != nil {
		responseError(res, fmt.Errorf("deleteUserUrlsHandler: decoding request: %w", err))
		return
	}
	user := req.Context().Value(config.UserCtxKey)
	if user == nil {
		responseError(res, errors.New("CheckIsUserURL, no user in context"))
		return
	}

	userID, ok := user.(string)
	if !ok {
		responseError(res, errors.New("CheckIsUserURL, wrong user type in context"))
		return
	}
	go s.cutter.DeleteUrls(userID, ids)
	res.WriteHeader(http.StatusAccepted)
}

func responseError(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(err.Error()))
}
