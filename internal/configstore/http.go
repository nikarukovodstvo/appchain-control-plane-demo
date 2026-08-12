package configstore

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) http.Handler {
	h := &Handler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/chain-configs", h.create)
	mux.HandleFunc("GET /v1/chain-configs/{id}", h.get)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return securityHeaders(mux)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request must be one JSON object with known fields")
		return
	}
	config, replayed, err := h.service.Create(r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrIdempotencyConflict):
			writeError(w, http.StatusConflict, "idempotency_conflict", err.Error())
		case errors.Is(err, ErrInvalidInput):
			writeError(w, http.StatusUnprocessableEntity, "invalid_input", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
		}
		return
	}
	if replayed {
		w.Header().Set("Idempotency-Replayed", "true")
		writeJSON(w, http.StatusOK, config)
		return
	}
	writeJSON(w, http.StatusCreated, config)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	config, err := h.service.Get(strings.TrimSpace(r.PathValue("id")))
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

