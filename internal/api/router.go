package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(handler *Handler, logger *slog.Logger) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(LoggingMiddleware(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	router.Get("/health", handler.HealthCheck)

	router.Route("/v1", func(apiRouter chi.Router) {
		apiRouter.Post("/check", handler.CheckCountry)
	})

	return router
}

func LoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			startTime := time.Now()

			wrappedWriter := &responseWriter{ResponseWriter: writer, statusCode: http.StatusOK}

			next.ServeHTTP(wrappedWriter, request)

			duration := time.Since(startTime)

			logger.Info("HTTP request",
				"method", request.Method,
				"path", request.URL.Path,
				"status", wrappedWriter.statusCode,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", request.RemoteAddr,
				"user_agent", request.UserAgent(),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (wrapper *responseWriter) WriteHeader(code int) {
	wrapper.statusCode = code
	wrapper.ResponseWriter.WriteHeader(code)
}
