package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// ListProducts godoc
// @Summary Список товаров
// @Tags products
// @Produce json
// @Success 200 {array} Product
// @Router /api/products [get]
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, name, description, price, created_at FROM products ORDER BY id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list products")
		return
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read products")
			return
		}
		products = append(products, p)
	}

	writeJSON(w, http.StatusOK, products)
}

// GetProduct godoc
// @Summary Получить товар по id
// @Tags products
// @Produce json
// @Param id path int true "ID товара"
// @Success 200 {object} Product
// @Failure 404 {object} ErrorResponse
// @Router /api/products/{id} [get]
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	p, err := h.fetchProduct(r, id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch product")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) fetchProduct(r *http.Request, id int64) (Product, error) {
	var p Product
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, name, description, price, created_at FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt)
	return p, err
}

// CreateProduct godoc
// @Summary Создать товар
// @Tags products
// @Accept json
// @Produce json
// @Param request body ProductRequest true "Данные товара"
// @Success 201 {object} Product
// @Failure 400 {object} ErrorResponse
// @Router /api/products [post]
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Price < 0 {
		writeError(w, http.StatusBadRequest, "name is required and price must be non-negative")
		return
	}

	var p Product
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO products (name, description, price) VALUES ($1, $2, $3)
		 RETURNING id, name, description, price, created_at`,
		req.Name, req.Description, req.Price,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create product")
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

// UpdateProduct godoc
// @Summary Обновить товар
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "ID товара"
// @Param request body ProductRequest true "Данные товара"
// @Success 200 {object} Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/products/{id} [put]
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Price < 0 {
		writeError(w, http.StatusBadRequest, "name is required and price must be non-negative")
		return
	}

	var p Product
	err = h.db.QueryRowContext(r.Context(),
		`UPDATE products SET name = $1, description = $2, price = $3 WHERE id = $4
		 RETURNING id, name, description, price, created_at`,
		req.Name, req.Description, req.Price, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update product")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

// DeleteProduct godoc
// @Summary Удалить товар
// @Tags products
// @Param id path int true "ID товара"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /api/products/{id} [delete]
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	res, err := h.db.ExecContext(r.Context(), `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete product")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
