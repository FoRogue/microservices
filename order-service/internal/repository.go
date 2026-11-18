package internal

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, order Order) (Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (user_id, status, total_cents) VALUES ($1, $2, $3)
		RETURNING id, created_at`, order.UserID, order.Status, order.Total).
		Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return Order{}, err
	}
	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price_cents)
			VALUES ($1, $2, $3, $4, $5)`, order.ID, item.ProductID, item.ProductName, item.Quantity, item.UnitPrice)
		if err != nil {
			return Order{}, err
		}
	}
	return order, tx.Commit()
}

func (r *Repository) Get(ctx context.Context, userID int64, orderID string) (Order, error) {
	order := Order{ID: orderID, UserID: userID}
	err := r.db.QueryRowContext(ctx, `
		SELECT status, total_cents, payment_id, created_at FROM orders
		WHERE id = $1 AND user_id = $2`, orderID, userID).
		Scan(&order.Status, &order.Total, &order.PaymentID, &order.CreatedAt)
	if err != nil {
		return Order{}, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT product_id, product_name, quantity, unit_price_cents
		FROM order_items WHERE order_id = $1 ORDER BY product_id`, orderID)
	if err != nil {
		return Order{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Quantity, &item.UnitPrice); err != nil {
			return Order{}, err
		}
		order.Items = append(order.Items, item)
	}
	return order, rows.Err()
}

func (r *Repository) RecordPayment(ctx context.Context, order Order, status string, paymentID int64) (Order, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE orders SET status = $1, payment_id = $2
		WHERE id = $3 AND user_id = $4 AND status = $5`,
		status, paymentID, order.ID, order.UserID, pending)
	if err != nil {
		return Order{}, err
	}
	return r.Get(ctx, order.UserID, order.ID)
}
