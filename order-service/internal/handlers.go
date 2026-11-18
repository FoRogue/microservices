package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	service *Service
	clients *Clients
}

func NewHandler(service *Service, clients *Clients) *Handler {
	return &Handler{service: service, clients: clients}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/orders", h.Create)
	mux.HandleFunc("GET /api/orders/{id}", h.Get)
	mux.HandleFunc("POST /api/orders/{id}/pay", h.Pay)
	return mux
}

// Create godoc
// @Summary Create an order from catalog products
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateOrderRequest true "Products and quantities"
// @Success 201 {object} Order
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/orders [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authorize(w, r)
	if !ok {
		return
	}
	var req CreateOrderRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{"invalid request body"})
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{"invalid request body"})
		return
	}
	order, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

// Get godoc
// @Summary Read your order and its status
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order UUID"
// @Success 200 {object} Order
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/orders/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, id, ok := h.orderRequest(w, r)
	if !ok {
		return
	}
	order, err := h.service.Get(r.Context(), userID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

// Pay godoc
// @Summary Pay for your order
// @Description Repeat the request if the payment service is temporarily unavailable.
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order UUID"
// @Success 200 {object} Order
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/orders/{id}/pay [post]
func (h *Handler) Pay(w http.ResponseWriter, r *http.Request) {
	userID, id, ok := h.orderRequest(w, r)
	if !ok {
		return
	}
	order, err := h.service.Pay(r.Context(), userID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *Handler) authorize(w http.ResponseWriter, r *http.Request) (int64, bool) {
	token := r.Header.Get("Authorization")
	if !strings.HasPrefix(token, "Bearer ") || len(token) <= len("Bearer ") {
		writeError(w, errUnauthorized)
		return 0, false
	}
	userID, err := h.clients.UserID(r.Context(), token)
	if err != nil {
		writeError(w, err)
		return 0, false
	}
	return userID, true
}

func (h *Handler) orderRequest(w http.ResponseWriter, r *http.Request) (int64, string, bool) {
	userID, ok := h.authorize(w, r)
	if !ok {
		return 0, "", false
	}
	var id pgtype.UUID
	if err := id.Scan(r.PathValue("id")); err != nil || !id.Valid {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{"invalid order id"})
		return 0, "", false
	}
	return userID, id.String(), true
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errUnauthorized):
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{err.Error()})
	case errors.Is(err, errInvalidItems), errors.Is(err, errInvalidTotal):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{err.Error()})
	case errors.Is(err, errUnknownProduct):
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{err.Error()})
	case errors.Is(err, errDependency):
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{err.Error()})
	case errors.Is(err, sql.ErrNoRows):
		writeJSON(w, http.StatusNotFound, ErrorResponse{"order not found"})
	default:
		log.Printf("order request failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{"order request failed"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
