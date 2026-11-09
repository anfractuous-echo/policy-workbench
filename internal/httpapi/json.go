package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"policy-workbench/internal/domain"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		writeResult(w, nil, domain.InvalidRequest("send application/json"), 0)
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeResult(w, nil, domain.InvalidRequest("invalid JSON: %v", err), 0)
		return false
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		writeResult(w, nil, domain.InvalidRequest("send one JSON object"), 0)
		return false
	}
	return true
}

func writeResult(w http.ResponseWriter, data any, err error, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		status, message := http.StatusInternalServerError, "Could not save the change. Check the server output and retry."
		switch {
		case errors.Is(err, domain.ErrInvalidRequest):
			status, message = http.StatusBadRequest, err.Error()
		case errors.Is(err, domain.ErrNotFound):
			status, message = http.StatusNotFound, "Not found in this profile. Refresh and retry."
		case errors.Is(err, domain.ErrActionUnavailable):
			status, message = http.StatusConflict, err.Error()
		default:
			log.Printf("request failed: %v", err)
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
		return
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
