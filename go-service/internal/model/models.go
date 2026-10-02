package model

import "time"

type Case struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Datetime    time.Time `json:"datetime"`
	Criticality string    `json:"criticality"`
	Description string    `json:"description"`
	AssignTo    *string   `json:"assign_to,omitempty"`
	Status      string    `json:"status"`
	Resolution  *string   `json:"resolution,omitempty"`
}

type Order struct {
	UUID        string    `json:"uuid"`
	CustomerID  string    `json:"customer_id"`
	ProductName string    `json:"product_name"`
	Amount      int64     `json:"amount"`
	Status      string    `json:"status"`
	OrderDate   time.Time `json:"order_date"`
	TrackingID  *string   `json:"tracking_id,omitempty"`
}

type User struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Division string `json:"division"`
}

type Doc struct {
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Doc       string    `json:"doc"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
