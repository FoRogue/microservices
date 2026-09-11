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

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at,
		       p.id, p.name, p.description, p.price, p.created_at
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.user_id = $1
		ORDER BY ci.id`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load cart")
		return
	}
	defer rows.Close()

	items := []CartItem{}
	for rows.Next() {
		var item CartItem
		var p Product
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt,
			&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read cart")
			return
		}
		item.Product = &p
		items = append(items, item)
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

	if _, err := h.fetchProduct(r, req.ProductID); errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to verify product")
		return
	}

	var item CartItem
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)
		 RETURNING id, user_id, product_id, quantity, created_at`,
		req.UserID, req.ProductID, req.Quantity,
	).Scan(&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt)
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

	var item CartItem
	err = h.db.QueryRowContext(r.Context(),
		`UPDATE cart_items SET quantity = $1 WHERE id = $2
		 RETURNING id, user_id, product_id, quantity, created_at`,
		req.Quantity, id,
	).Scan(&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt)
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

	res, err := h.db.ExecContext(r.Context(), `DELETE FROM cart_items WHERE id = $1`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete cart item")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "cart item not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
