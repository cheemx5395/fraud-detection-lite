package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		// Route template (better than raw path)
		route := "unknown"
		if currentRoute := mux.CurrentRoute(r); currentRoute != nil {
			if tmpl, err := currentRoute.GetPathTemplate(); err == nil {
				route = tmpl
			}
		}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		log.Printf(
			"%s %s status: %d duration: %s\n",
			r.Method,
			route,
			rec.status,
			duration,
		)
	})
}
