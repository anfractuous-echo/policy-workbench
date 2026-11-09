package httpapi

import (
	"io/fs"
	"net/http"
	"strings"

	"policy-workbench/internal/application"
	"policy-workbench/internal/domain"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service, assets fs.FS, brief []byte) http.Handler {
	handler := &Handler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", handler.workspace)
	mux.HandleFunc("POST /api/policies", handler.createPolicy)
	mux.HandleFunc("DELETE /api/policies/{id}", handler.deletePolicy)
	mux.HandleFunc("PUT /api/policies/order", handler.reorderPolicies)
	mux.HandleFunc("POST /api/transfers", handler.createTransfer)
	mux.HandleFunc("POST /api/transfers/{id}/actions", handler.applyTransferAction)
	mux.HandleFunc("POST /api/clock", handler.advanceClock)
	mux.HandleFunc("POST /api/reset", handler.resetWorkspace)
	mux.HandleFunc("GET /brief", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write(brief)
	})
	mux.Handle("GET /", http.FileServer(http.FS(assets)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; object-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
			if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				writeResult(w, nil, domain.InvalidRequest("cross-site API requests are not supported"), 0)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}
