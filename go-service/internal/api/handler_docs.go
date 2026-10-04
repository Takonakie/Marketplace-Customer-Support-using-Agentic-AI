package api

import (
	"encoding/json"
	"net/http"

	"github.com/agenticsdk/go-service/internal/model"
	"github.com/agenticsdk/go-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

type DocHandler struct {
	repo *repository.DocRepo
}

func NewDocHandler(repo *repository.DocRepo) *DocHandler {
	return &DocHandler{repo: repo}
}

func (h *DocHandler) ListDocs(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		http.Error(w, `{"error":"Database not available"}`, http.StatusServiceUnavailable)
		return
	}

	docs, err := h.repo.ListDocs(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if docs == nil {
		docs = []model.Doc{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

type CreateDocRequest struct {
	Name string `json:"name"`
	Doc  string `json:"doc"`
}

func (h *DocHandler) CreateDoc(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		http.Error(w, `{"error":"Database not available"}`, http.StatusServiceUnavailable)
		return
	}

	var req CreateDocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Doc == "" {
		http.Error(w, `{"error":"name and doc content are required"}`, http.StatusBadRequest)
		return
	}

	doc, err := h.repo.CreateDoc(r.Context(), req.Name, req.Doc)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

func (h *DocHandler) DeleteDoc(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		http.Error(w, `{"error":"Database not available"}`, http.StatusServiceUnavailable)
		return
	}

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		http.Error(w, `{"error":"uuid is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteDoc(r.Context(), uuid); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"deleted"}`))
}

func (h *DocHandler) UpdateDoc(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		http.Error(w, `{"error":"Database not available"}`, http.StatusServiceUnavailable)
		return
	}

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		http.Error(w, `{"error":"uuid is required"}`, http.StatusBadRequest)
		return
	}

	var req CreateDocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Doc == "" {
		http.Error(w, `{"error":"name and doc content are required"}`, http.StatusBadRequest)
		return
	}

	doc, err := h.repo.UpdateDoc(r.Context(), uuid, req.Name, req.Doc)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(doc)
}
