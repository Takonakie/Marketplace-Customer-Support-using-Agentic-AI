package api

import (
	"net/http"

	"github.com/agenticsdk/go-service/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(orderRepo *repository.OrderRepo, caseRepo *repository.CaseRepo, userRepo *repository.UserRepo, docRepo *repository.DocRepo) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	orderH := NewOrderHandler(orderRepo)
	caseH := NewCaseHandler(caseRepo, userRepo)
	userH := NewUserHandler(userRepo)

	r.Route("/api", func(r chi.Router) {
		r.Get("/orders", orderH.GetOrders)
		r.Get("/orders/total", orderH.GetTotal)

		r.Post("/cases", caseH.CreateCase)
		r.Get("/cases/{uuid}", caseH.GetCase)
		r.Patch("/cases/{uuid}/status", caseH.UpdateCaseStatus)

		r.Get("/users", userH.GetUsers)
	})

	return r
}
