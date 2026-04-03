package handler

import (
	"ims-database-util/internal/service"
	"ims-database-util/internal/utils"
	"net/http"
)

type UserHandler struct {
	service service.UserService
}

// NewUserHandler creates a UserHandler that uses the provided UserService.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-user-id")
	if userId == "" {
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing user ID header"})
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userId)
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
