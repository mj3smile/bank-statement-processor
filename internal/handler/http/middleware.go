package http

import (
	"net/http"

	"github.com/mj3smile/bank-statement-processor/internal/infra/log"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, logID := log.InjectNewID(r.Context())
		r = r.WithContext(ctx)

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		lrw.Header().Set("X-Log-ID", logID)

		next.ServeHTTP(lrw, r)
	})
}
