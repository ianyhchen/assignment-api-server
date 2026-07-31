package httpapi

import "net/http"

// NewRouter registers the task routes and returns their HTTP handler.
func NewRouter(handler *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", handler.List)
	mux.HandleFunc("POST /tasks", handler.Create)
	mux.HandleFunc("PUT /tasks/{id}", handler.Update)
	mux.HandleFunc("DELETE /tasks/{id}", handler.Delete)

	return mux
}
