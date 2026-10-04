package repository

import (
	"context"

	"github.com/agenticsdk/go-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) ListAllUsers(ctx context.Context) ([]model.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT uuid, name, division FROM users ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.UUID, &u.Name, &u.Division); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) GetUsersByDivision(ctx context.Context, division string) ([]model.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT uuid, name, division FROM users WHERE division = $1`, division)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.UUID, &u.Name, &u.Division); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, name, division string) (*model.User, error) {
	var u model.User
	query := `INSERT INTO users (name, division) VALUES ($1, $2) RETURNING uuid, name, division`
	err := r.pool.QueryRow(ctx, query, name, division).Scan(&u.UUID, &u.Name, &u.Division)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, uuid string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE uuid = $1`, uuid)
	return err
}
