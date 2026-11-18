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
	products, err := h.repo.ListProducts(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list products")
		return
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

	p, err := h.repo.GetProduct(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch product")
		return
	}

	writeJSON(w, http.StatusOK, p)
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
	if req.Name == "" || req.Price.sign() < 0 {
		writeError(w, http.StatusBadRequest, "name is required and price must be non-negative")
		return
	}

	p, err := h.repo.CreateProduct(r.Context(), req)
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
	if req.Name == "" || req.Price.sign() < 0 {
		writeError(w, http.StatusBadRequest, "name is required and price must be non-negative")
		return
	}

	p, err := h.repo.UpdateProduct(r.Context(), id, req)
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

	err = h.repo.DeleteProduct(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
