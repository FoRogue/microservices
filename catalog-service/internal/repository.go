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

func (repo *Repository) ListProducts(ctx context.Context) ([]Product, error) {
	rows, err := repo.db.QueryContext(ctx,
		`SELECT id, name, description, price, created_at FROM products ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (repo *Repository) GetProduct(ctx context.Context, id int64) (Product, error) {
	var p Product
	err := repo.db.QueryRowContext(ctx,
		`SELECT id, name, description, price, created_at FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt)
	return p, err
}

func (repo *Repository) CreateProduct(ctx context.Context, req ProductRequest) (Product, error) {
	var p Product
	err := repo.db.QueryRowContext(ctx,
		`INSERT INTO products (name, description, price) VALUES ($1, $2, $3::text::numeric)
		 RETURNING id, name, description, price, created_at`,
		req.Name, req.Description, req.Price.databaseValue(),
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt)
	return p, err
}

func (repo *Repository) UpdateProduct(ctx context.Context, id int64, req ProductRequest) (Product, error) {
	var p Product
	err := repo.db.QueryRowContext(ctx,
		`UPDATE products SET name = $1, description = $2, price = $3::text::numeric WHERE id = $4
		 RETURNING id, name, description, price, created_at`,
		req.Name, req.Description, req.Price.databaseValue(), id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.CreatedAt)
	return p, err
}

func (repo *Repository) DeleteProduct(ctx context.Context, id int64) error {
	res, err := repo.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *Repository) GetCart(ctx context.Context, userID string) ([]CartItem, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at,
		       p.id, p.name, p.description, p.price, p.created_at
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.user_id = $1
		ORDER BY ci.id`, userID)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		item.Product = &p
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (repo *Repository) AddCartItem(ctx context.Context, req AddCartItemRequest) (CartItem, error) {
	var item CartItem
	err := repo.db.QueryRowContext(ctx,
		`INSERT INTO cart_items (user_id, product_id, quantity)
		 SELECT $1, id, $3 FROM products WHERE id = $2 FOR KEY SHARE
		 RETURNING id, user_id, product_id, quantity, created_at`,
		req.UserID, req.ProductID, req.Quantity,
	).Scan(&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt)
	return item, err
}

func (repo *Repository) UpdateCartItem(ctx context.Context, id int64, quantity int) (CartItem, error) {
	var item CartItem
	err := repo.db.QueryRowContext(ctx,
		`UPDATE cart_items SET quantity = $1 WHERE id = $2
		 RETURNING id, user_id, product_id, quantity, created_at`,
		quantity, id,
	).Scan(&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt)
	return item, err
}

func (repo *Repository) DeleteCartItem(ctx context.Context, id int64) error {
	res, err := repo.db.ExecContext(ctx, `DELETE FROM cart_items WHERE id = $1`, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}
