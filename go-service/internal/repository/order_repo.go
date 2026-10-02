package repository

import (
	"context"

	"github.com/agenticsdk/go-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

func (r *OrderRepo) GetOrdersByCustomerID(ctx context.Context, customerID string) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx, `SELECT uuid, customer_id, product_name, amount, status, order_date, tracking_id FROM orders WHERE customer_id = $1`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.UUID, &o.CustomerID, &o.ProductName, &o.Amount, &o.Status, &o.OrderDate, &o.TrackingID); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrderRepo) GetOrderByUUID(ctx context.Context, uuid string) (*model.Order, error) {
	var o model.Order
	err := r.pool.QueryRow(ctx, `SELECT uuid, customer_id, product_name, amount, status, order_date, tracking_id FROM orders WHERE uuid = $1`, uuid).
		Scan(&o.UUID, &o.CustomerID, &o.ProductName, &o.Amount, &o.Status, &o.OrderDate, &o.TrackingID)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) GetTotalPurchaseByCustomerID(ctx context.Context, customerID string) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM orders WHERE customer_id = $1 AND status != 'cancelled'`, customerID).Scan(&total)
	return total, err
}
