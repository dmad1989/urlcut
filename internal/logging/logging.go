package logging

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func Initilize() error {
	zl, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("Logger.Initlogs: %w", err)
	}
	Log = zl
	return nil
}

func WithLog(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		Log.Sugar().Infoln(
			"uri:", uri,
			"; method:", method,
			"; duration:", duration,
			"; status:", responseData.status,
			"; size:", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
