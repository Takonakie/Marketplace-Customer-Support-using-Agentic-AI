package repository

import (
	"context"

	"github.com/agenticsdk/go-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CaseRepo struct {
	pool *pgxpool.Pool
}

func NewCaseRepo(pool *pgxpool.Pool) *CaseRepo {
	return &CaseRepo{pool: pool}
}

func (r *CaseRepo) CreateCase(ctx context.Context, c *model.Case) error {
	query := `INSERT INTO cases (name, criticality, description, assign_to, status) VALUES ($1, $2, $3, $4, $5) RETURNING uuid, datetime`
	return r.pool.QueryRow(ctx, query, c.Name, c.Criticality, c.Description, c.AssignTo, c.Status).Scan(&c.UUID, &c.Datetime)
}

func (r *CaseRepo) GetCaseByUUID(ctx context.Context, uuid string) (*model.Case, error) {
	var c model.Case
	err := r.pool.QueryRow(ctx, `SELECT uuid, name, datetime, criticality, description, assign_to, status, resolution FROM cases WHERE uuid = $1`, uuid).
		Scan(&c.UUID, &c.Name, &c.Datetime, &c.Criticality, &c.Description, &c.AssignTo, &c.Status, &c.Resolution)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CaseRepo) UpdateCaseStatus(ctx context.Context, uuid, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE cases SET status = $1, updated_at = NOW() WHERE uuid = $2`, status, uuid)
	return err
}
