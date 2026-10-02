package repository

import (
	"context"

	"github.com/agenticsdk/go-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocRepo struct {
	pool *pgxpool.Pool
}

func NewDocRepo(pool *pgxpool.Pool) *DocRepo {
	return &DocRepo{pool: pool}
}

func (r *DocRepo) GetDocByName(ctx context.Context, name string) (*model.Doc, error) {
	var d model.Doc
	err := r.pool.QueryRow(ctx, `SELECT uuid, name, doc, created_at, updated_at FROM docs WHERE name ILIKE $1`, "%"+name+"%").
		Scan(&d.UUID, &d.Name, &d.Doc, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
