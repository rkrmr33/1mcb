package util

import (
	"fmt"
	"net/http"
	"time"
)

type respWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

var _ http.ResponseWriter = &respWriterWrapper{}
var _ http.Flusher = &respWriterWrapper{}

func (w *respWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(w.statusCode)
}

// Flush implements http.Flusher.
func (w *respWriterWrapper) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func LoggerMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Printf("%s --> %s %s %s\n", start.Format(time.RFC3339), r.Method, r.URL.Path, r.RemoteAddr)

		wrappedRespWriter := respWriterWrapper{w, 200}
		handler.ServeHTTP(&wrappedRespWriter, r)

		end := time.Now()
		fmt.Printf("%s <-- %s %s %s %d %s\n", end.Format(time.RFC3339), r.Method, r.URL.Path, r.RemoteAddr, wrappedRespWriter.statusCode, time.Duration(end.UnixNano()-start.UnixNano()))
	})
}
