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

func (repo *Repository) CreatePayment(ctx context.Context, req CreatePaymentRequest, status string) (Payment, error) {
	var p Payment
	err := repo.db.QueryRowContext(ctx,
		`INSERT INTO payments (order_id, amount, status) VALUES ($1, $2::text::numeric, $3)
		 RETURNING id, order_id, amount, status, created_at`,
		req.OrderID, req.Amount.databaseValue(), status,
	).Scan(&p.ID, &p.OrderID, &p.Amount, &p.Status, &p.CreatedAt)
	return p, err
}

func (repo *Repository) GetPayment(ctx context.Context, id int64) (Payment, error) {
	var p Payment
	err := repo.db.QueryRowContext(ctx,
		`SELECT id, order_id, amount, status, created_at FROM payments WHERE id = $1`, id,
	).Scan(&p.ID, &p.OrderID, &p.Amount, &p.Status, &p.CreatedAt)
	return p, err
}

func (repo *Repository) ListPayments(ctx context.Context, orderID string) ([]Payment, error) {
	var rows *sql.Rows
	var err error
	if orderID != "" {
		rows, err = repo.db.QueryContext(ctx,
			`SELECT id, order_id, amount, status, created_at FROM payments WHERE order_id = $1 ORDER BY id`, orderID)
	} else {
		rows, err = repo.db.QueryContext(ctx,
			`SELECT id, order_id, amount, status, created_at FROM payments ORDER BY id`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := []Payment{}
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.Amount, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payments, nil
}
