package api

import (
	"encoding/json"
	"net/http"

	"github.com/agenticsdk/go-service/internal/model"
	"github.com/agenticsdk/go-service/internal/repository"
)

type OrderHandler struct {
	repo *repository.OrderRepo
}

func NewOrderHandler(repo *repository.OrderRepo) *OrderHandler {
	return &OrderHandler{repo: repo}
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	var orders []model.Order
	var err error

	if customerID == "" {
		orders, err = h.repo.ListAllOrders(r.Context())
	} else {
		orders, err = h.repo.GetOrdersByCustomerID(r.Context(), customerID)
	}

	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if orders == nil {
		orders = []model.Order{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetTotal(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	if customerID == "" {
		http.Error(w, `{"error":"customer_id is required"}`, http.StatusBadRequest)
		return
	}

	total, err := h.repo.GetTotalPurchaseByCustomerID(r.Context(), customerID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"customer_id":    customerID,
		"total_purchase": total,
	})
}

type CreateOrderRequest struct {
	CustomerID  string `json:"customer_id"`
	ProductName string `json:"product_name"`
	Amount      int64  `json:"amount"`
	Status      string `json:"status"`
	TrackingID  string `json:"tracking_id"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		http.Error(w, `{"error":"Database not available"}`, http.StatusServiceUnavailable)
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.CustomerID == "" || req.ProductName == "" {
		http.Error(w, `{"error":"customer_id and product_name are required"}`, http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = "shipped"
	}

	var trackingPtr *string
	if req.TrackingID != "" {
		trackingPtr = &req.TrackingID
	}

	order := &model.Order{
		CustomerID:  req.CustomerID,
		ProductName: req.ProductName,
		Amount:      req.Amount,
		Status:      req.Status,
		TrackingID:  trackingPtr,
	}

	if err := h.repo.CreateOrder(r.Context(), order); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
