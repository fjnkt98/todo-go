package server

import (
	"log/slog"
	"net/http"
	"text/template"
)

type IndexHandler struct{}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFS(templates, "templates/index.html", "templates/base.html")
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), "parse template", slog.Any("error", err))
		return
	}

	type Data struct {
		Title   string
		Content string
	}

	data := Data{
		Title:   "ToDo Go",
		Content: "This is my first html/template content.",
	}

	w.WriteHeader(http.StatusOK)
	if err := t.Execute(w, &data); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), "write response", slog.Any("error", err))
		return
	}
}
