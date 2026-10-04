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
	if customerID == "" {
		http.Error(w, `{"error":"customer_id is required"}`, http.StatusBadRequest)
		return
	}

	orders, err := h.repo.GetOrdersByCustomerID(r.Context(), customerID)
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
