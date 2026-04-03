package handler

import (
	"ims-database-util/internal/repository"
	"ims-database-util/internal/utils"
	"log/slog"
	"net/http"
)

type ProductHandler struct {
	repo repository.ProductRepository
}

func NewProductHandler(repo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (h *ProductHandler) GetProductByUserId(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-user-id")
	if userId == "" {
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing user ID header"})
		return
	}

	products, err := h.repo.GetProductsByUserId(r.Context(), userId)
	if err != nil {
		slog.Error("Error Occured", "ERROR", err)
		utils.JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Error in fetching products"})
		return
	}

	utils.JSONResponse(w, http.StatusOK, products)

}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.repo.GetProducts(r.Context())
	if err != nil {
		slog.Error("Error Occured", "ERROR", err)
		utils.JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Error in fetching products"})
		return
	}

	utils.JSONResponse(w, http.StatusOK, products)
}
