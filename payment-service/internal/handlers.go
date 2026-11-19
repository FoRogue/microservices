package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /api/payments", h.CreatePayment)
	mux.HandleFunc("GET /api/payments", h.ListPayments)
	mux.HandleFunc("GET /api/payments/{id}", h.GetPayment)
	return mux
}

// Health godoc
// @Summary Проверка живости сервиса
// @Tags health
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreatePayment godoc
// @Summary Провести оплату заказа
// @Description Симулирует платёжный шлюз: положительная сумма — успех, иначе отказ.
// @Tags payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "Данные платежа"
// @Success 201 {object} Payment
// @Failure 400 {object} ErrorResponse
// @Param Idempotency-Key header string false "Reuse the key when retrying the same payment"
// @Failure 409 {object} ErrorResponse
// @Router /api/payments [post]
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.OrderID == "" {
		writeError(w, http.StatusBadRequest, "order_id is required")
		return
	}

	key := r.Header.Get("Idempotency-Key")
	if strings.HasPrefix(key, "kafka-order:") {
		writeError(w, http.StatusBadRequest, "idempotency key uses a reserved prefix")
		return
	}
	if len(key) > 128 {
		writeError(w, http.StatusBadRequest, "idempotency key is too long")
		return
	}
	p, err := h.service.CreatePayment(r.Context(), req, key)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusConflict, "idempotency key already used for another payment")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create payment")
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

// GetPayment godoc
// @Summary Получить платёж по id
// @Tags payments
// @Produce json
// @Param id path int true "ID платежа"
// @Success 200 {object} Payment
// @Failure 404 {object} ErrorResponse
// @Router /api/payments/{id} [get]
func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment id")
		return
	}

	p, err := h.service.GetPayment(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "payment not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch payment")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

// ListPayments godoc
// @Summary Список платежей, опционально по заказу
// @Tags payments
// @Produce json
// @Param order_id query string false "ID заказа"
// @Success 200 {array} Payment
// @Router /api/payments [get]
func (h *Handler) ListPayments(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")

	payments, err := h.service.ListPayments(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list payments")
		return
	}

	writeJSON(w, http.StatusOK, payments)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}
