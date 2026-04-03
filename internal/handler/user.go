package handler

import (
	"ims-database-util/internal/repository"
	"ims-database-util/internal/utils"
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
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing user ID header"})
		return
	}
	user, err := h.repo.GetUserByID(r.Context(), userId)

	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	if user == nil {
		utils.JSONResponse(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	utils.JSONResponse(w, http.StatusOK, user)
}
