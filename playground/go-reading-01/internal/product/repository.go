package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByID(ctx context.Context, id int64) (Product, error) {
	if id <= 0 {
		return Product{}, &ValidationError{Field: "id", Reason: "must be positive"}
	}

	var p Product
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, price, stock FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("find product %d: %w", id, err)
	}
	return p, nil
}
