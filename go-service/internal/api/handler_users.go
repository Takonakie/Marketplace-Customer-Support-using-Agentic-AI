package api

import (
	"encoding/json"
	"net/http"

	"github.com/agenticsdk/go-service/internal/model"
	"github.com/agenticsdk/go-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	repo *repository.UserRepo
}

func NewUserHandler(repo *repository.UserRepo) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	div := r.URL.Query().Get("division")
	var users []model.User
	var err error

	if div != "" {
		users, err = h.repo.GetUsersByDivision(r.Context(), div)
	} else {
		users, err = h.repo.ListAllUsers(r.Context())
	}

	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []model.User{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

type CreateUserReq struct {
	Name     string `json:"name"`
	Division string `json:"division"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Division == "" {
		http.Error(w, `{"error":"name and division are required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.repo.CreateUser(r.Context(), req.Name, req.Division)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		http.Error(w, `{"error":"uuid is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteUser(r.Context(), uuid); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"deleted"}`))
}
