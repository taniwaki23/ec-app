package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"example.com/go-reading-01/internal/product"
)

type Handler struct {
	db       *sql.DB
	products *product.Service
}

func NewHandler(db *sql.DB, products *product.Service) *Handler {
	return &Handler{db: db, products: products}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /products/{id}", h.getProduct)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if err := h.db.PingContext(r.Context()); err != nil {
		log.Printf("health: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id must be an integer"})
		return
	}

	p, err := h.products.Get(r.Context(), id)
	if err != nil {
		var verr *product.ValidationError
		switch {
		case errors.As(err, &verr):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": verr.Error()})
		case errors.Is(err, product.ErrNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
		case errors.Is(err, context.DeadlineExceeded):
			writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "request timed out"})

		default:
			log.Printf("get product: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}
