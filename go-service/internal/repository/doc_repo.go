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

func (r *DocRepo) ListDocs(ctx context.Context) ([]model.Doc, error) {
	rows, err := r.pool.Query(ctx, `SELECT uuid, name, doc, created_at, updated_at FROM docs ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []model.Doc
	for rows.Next() {
		var d model.Doc
		if err := rows.Scan(&d.UUID, &d.Name, &d.Doc, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}

func (r *DocRepo) CreateDoc(ctx context.Context, name, docContent string) (*model.Doc, error) {
	var d model.Doc
	query := `INSERT INTO docs (name, doc, embedding) VALUES ($1, $2, NULL) RETURNING uuid, name, doc, created_at, updated_at`
	err := r.pool.QueryRow(ctx, query, name, docContent).Scan(&d.UUID, &d.Name, &d.Doc, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DocRepo) DeleteDoc(ctx context.Context, uuid string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM docs WHERE uuid = $1`, uuid)
	return err
}
