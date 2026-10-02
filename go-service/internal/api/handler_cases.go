package api

import (
	"encoding/json"
	"net/http"

	"github.com/agenticsdk/go-service/internal/model"
	"github.com/agenticsdk/go-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

type CaseHandler struct {
	caseRepo *repository.CaseRepo
	userRepo *repository.UserRepo
}

func NewCaseHandler(caseRepo *repository.CaseRepo, userRepo *repository.UserRepo) *CaseHandler {
	return &CaseHandler{caseRepo: caseRepo, userRepo: userRepo}
}

type CreateCaseReq struct {
	Name             string `json:"name"`
	Criticality      string `json:"criticality"`
	Description      string `json:"description"`
	AssignToDivision string `json:"assign_to_division"`
}

type UpdateCaseStatusReq struct {
	Status     string `json:"status"`
	Resolution string `json:"resolution,omitempty"`
}

func (h *CaseHandler) CreateCase(w http.ResponseWriter, r *http.Request) {
	var req CreateCaseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	var assignTo *string
	if req.AssignToDivision != "" {
		users, err := h.userRepo.GetUsersByDivision(r.Context(), req.AssignToDivision)
		if err == nil && len(users) > 0 {
			assignTo = &users[0].UUID
		}
	}

	newCase := model.Case{
		Name:        req.Name,
		Criticality: req.Criticality,
		Description: req.Description,
		AssignTo:    assignTo,
		Status:      "open",
	}

	if err := h.caseRepo.CreateCase(r.Context(), &newCase); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCase)
}

func (h *CaseHandler) GetCase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "uuid")
	c, err := h.caseRepo.GetCaseByUUID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"case not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *CaseHandler) UpdateCaseStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "uuid")
	var req UpdateCaseStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		http.Error(w, `{"error":"status is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.caseRepo.UpdateCaseStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, `{"error":"failed to update status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"uuid":   id,
		"status": req.Status,
		"msg":    "Case status updated successfully",
	})
}
