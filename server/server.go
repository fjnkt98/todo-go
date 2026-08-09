package server

import (
	"embed"
	"fmt"
	"net/http"
	"time"
)

//go:embed templates
var templates embed.FS

//go:embed static
var statics embed.FS

func NewServer(port int) (*http.Server, error) {
	mux, err := NewServeMux()
	if err != nil {
		return nil, fmt.Errorf("create new serve mux: %w", err)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      RecoveryMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return server, nil
}

func NewServeMux() (*http.ServeMux, error) {
	mux := http.NewServeMux()
	mux.Handle("GET /", &IndexHandler{})
	mux.Handle("GET /static/", http.FileServer(http.FS(statics)))
	return mux, nil
}
