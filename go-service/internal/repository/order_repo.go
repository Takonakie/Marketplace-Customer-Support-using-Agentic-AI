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
	query := `SELECT uuid, customer_id, product_name, amount, status, order_date, tracking_id 
	          FROM orders 
	          WHERE customer_id = $1 OR customer_id = REPLACE($1, 'tg_', '') OR customer_id = 'tg_' || $1`
	rows, err := r.pool.Query(ctx, query, customerID)
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
	query := `SELECT COALESCE(SUM(amount), 0) 
	          FROM orders 
	          WHERE (customer_id = $1 OR customer_id = REPLACE($1, 'tg_', '') OR customer_id = 'tg_' || $1) 
	            AND status != 'cancelled'`
	err := r.pool.QueryRow(ctx, query, customerID).Scan(&total)
	return total, err
}

func (r *OrderRepo) CreateOrder(ctx context.Context, o *model.Order) error {
	query := `INSERT INTO orders (customer_id, product_name, amount, status, tracking_id) VALUES ($1, $2, $3, $4, $5) RETURNING uuid, order_date`
	return r.pool.QueryRow(ctx, query, o.CustomerID, o.ProductName, o.Amount, o.Status, o.TrackingID).Scan(&o.UUID, &o.OrderDate)
}

func (r *OrderRepo) ListAllOrders(ctx context.Context) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx, `SELECT uuid, customer_id, product_name, amount, status, order_date, tracking_id FROM orders ORDER BY order_date DESC LIMIT 50`)
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
