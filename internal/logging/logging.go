// Package logging инициализирует логирование работы сервиса.
package logging

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Log глобальная перемнная для логирования.
var Log *zap.SugaredLogger = zap.NewNop().Sugar()

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
		wroteHeader  bool
	}
)

// Write добавляет в логирование запроса значение размера и статуса.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	if !r.wroteHeader {
		r.responseData.status = http.StatusOK
	}
	return size, err
}

// WriteHeader добавляет в логирование ответа значение статуса.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	if !r.wroteHeader {
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(statusCode)
		r.responseData.status = statusCode
	}
}

// Initilize создаем общий zap - логгер.
func Initilize() error {
	zl, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("Logger.Initlogs: %w", err)
	}
	Log = zl.Sugar()
	return nil
}

// WithLog - это middleware, отвечающая за логирование каждого запроса и ответа.
// Логируется URL, метод, время выполнения, статус ответа и размер.
func WithLog(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		Log.Infow("Request",
			zap.String("uri", uri),
			zap.String("method", method))
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		defer Log.Sync()

		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		Log.Infow("Response",
			zap.Duration("duration", duration),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)
	}
	return http.HandlerFunc(logFn)
}
