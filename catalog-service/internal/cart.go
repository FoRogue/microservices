package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// GetCart godoc
// @Summary Содержимое корзины пользователя
// @Tags cart
// @Produce json
// @Param user_id query string true "ID пользователя"
// @Success 200 {array} CartItem
// @Failure 400 {object} ErrorResponse
// @Router /api/cart [get]
func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}

	items, err := h.repo.GetCart(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load cart")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// AddCartItem godoc
// @Summary Добавить товар в корзину
// @Tags cart
// @Accept json
// @Produce json
// @Param request body AddCartItemRequest true "Товар и количество"
// @Success 201 {object} CartItem
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/cart/items [post]
func (h *Handler) AddCartItem(w http.ResponseWriter, r *http.Request) {
	var req AddCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" || req.ProductID <= 0 || req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "user_id, product_id and a positive quantity are required")
		return
	}

	item, err := h.repo.AddCartItem(r.Context(), req)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add item to cart")
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// UpdateCartItem godoc
// @Summary Изменить количество товара в корзине
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "ID позиции корзины"
// @Param request body UpdateCartItemRequest true "Новое количество"
// @Success 200 {object} CartItem
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/cart/items/{id} [put]
func (h *Handler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cart item id")
		return
	}

	var req UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "quantity must be positive")
		return
	}

	item, err := h.repo.UpdateCartItem(r.Context(), id, req.Quantity)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "cart item not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update cart item")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// DeleteCartItem godoc
// @Summary Удалить товар из корзины
// @Tags cart
// @Param id path int true "ID позиции корзины"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /api/cart/items/{id} [delete]
func (h *Handler) DeleteCartItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cart item id")
		return
	}

	err = h.repo.DeleteCartItem(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "cart item not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete cart item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
