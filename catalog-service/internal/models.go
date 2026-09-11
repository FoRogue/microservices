package internal

import "time"

// Product is an item available in the catalog.
type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProductRequest is the payload for creating/updating a product.
type ProductRequest struct {
	Name        string  `json:"name" example:"Клавиатура"`
	Description string  `json:"description" example:"Механическая клавиатура"`
	Price       float64 `json:"price" example:"4990.00"`
}

// CartItem is a product placed in a user's cart, with product details joined in.
type CartItem struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Product   *Product  `json:"product,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// AddCartItemRequest is the payload for POST /api/cart/items.
type AddCartItemRequest struct {
	UserID    string `json:"user_id" example:"user-1"`
	ProductID int64  `json:"product_id" example:"1"`
	Quantity  int    `json:"quantity" example:"2"`
}

// UpdateCartItemRequest is the payload for PUT /api/cart/items/{id}.
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" example:"3"`
}

// ErrorResponse is a generic error payload.
type ErrorResponse struct {
	Error string `json:"error" example:"product not found"`
}
