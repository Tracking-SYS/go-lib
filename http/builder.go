package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/lk153/go-lib/trace"
)

//NewHTTPBuilder ...
func NewHTTPBuilder() *trace.Builder {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowOriginFunc:  AllowTrackingService,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Use(NewHTTPMetricsObserver("tracking", trace.ShouldLogRequest))

	return &trace.Builder{
		Mux:  mux,
		Name: "tracking",
	}
}

func AllowTrackingService(_ *http.Request, origin string) bool {
	return false
}
