package api

import (
	"net/http"
	"os"
	"path/filepath"

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
	docH := NewDocHandler(docRepo)

	r.Route("/api", func(r chi.Router) {
		r.Get("/orders", orderH.GetOrders)
		r.Post("/orders", orderH.CreateOrder)
		r.Get("/orders/total", orderH.GetTotal)

		r.Get("/cases", caseH.ListCases)
		r.Post("/cases", caseH.CreateCase)
		r.Get("/cases/{uuid}", caseH.GetCase)
		r.Patch("/cases/{uuid}/status", caseH.UpdateCaseStatus)

		r.Get("/docs", docH.ListDocs)
		r.Post("/docs", docH.CreateDoc)
		r.Delete("/docs/{uuid}", docH.DeleteDoc)

		r.Get("/users", userH.GetUsers)
	})

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "static", "dashboard.html"))
	})
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(filesDir)))

	return r
}
