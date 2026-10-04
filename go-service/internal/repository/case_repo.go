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

func (r *CaseRepo) GetAllCases(ctx context.Context) ([]model.Case, error) {
	rows, err := r.pool.Query(ctx, `SELECT uuid, name, datetime, criticality, description, COALESCE(assign_to::text, ''), status, COALESCE(resolution, '') FROM cases ORDER BY datetime DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []model.Case
	for rows.Next() {
		var c model.Case
		var assignTo string
		if err := rows.Scan(&c.UUID, &c.Name, &c.Datetime, &c.Criticality, &c.Description, &assignTo, &c.Status, &c.Resolution); err != nil {
			return nil, err
		}
		if assignTo != "" {
			c.AssignTo = &assignTo
		}
		cases = append(cases, c)
	}
	return cases, nil
}

func (r *CaseRepo) GetCasesByCustomerID(ctx context.Context, customerID string) ([]model.Case, error) {
	cleanID := customerID
	if len(cleanID) > 3 && cleanID[:3] == "tg_" {
		cleanID = cleanID[3:]
	}
	query := `SELECT uuid, name, datetime, criticality, description, COALESCE(assign_to::text, ''), status, COALESCE(resolution, '') 
	          FROM cases 
	          WHERE description ILIKE $1 OR description ILIKE $2 
	          ORDER BY datetime DESC`
	rows, err := r.pool.Query(ctx, query, "%"+customerID+"%", "%"+cleanID+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []model.Case
	for rows.Next() {
		var c model.Case
		var assignTo string
		if err := rows.Scan(&c.UUID, &c.Name, &c.Datetime, &c.Criticality, &c.Description, &assignTo, &c.Status, &c.Resolution); err != nil {
			return nil, err
		}
		if assignTo != "" {
			c.AssignTo = &assignTo
		}
		cases = append(cases, c)
	}
	return cases, nil
}
