// Package server produces http server
package server

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//go:embed templates
var templates embed.FS

//go:embed static
var statics embed.FS

func NewServer(port int) (*http.Server, error) {
	mux := http.NewServeMux()

	mux.Handle("GET /", &IndexHandler{})
	mux.Handle("GET /static/", http.FileServer(http.FS(statics)))

	handler := RecoveryMiddleware(mux)
	handler = otelhttp.NewHandler(handler, "http-request")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return server, nil
}
