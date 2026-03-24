package handler

import (
	"encoding/json"
	"ims-database-util/internal/repository"
	"net/http"
)

type UserHandler struct {
	repo repository.UserRepository
}

// NewUserHandler creates a UserHandler that uses the provided repository for user operations.
func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-user-id")
	if userId == "" {
		h.jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing user ID header"})
		return
	}
	user, err := h.repo.GetUserByID(r.Context(), userId)

	if err != nil {
		h.jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	if user == nil {
		h.jsonResponse(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	h.jsonResponse(w, http.StatusOK, user)
}

func (h *UserHandler) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
