package handler

import (
	"encoding/json"
	"fmt"
	"ims-database-util/internal/repository"
	"ims-database-util/internal/service"
	"ims-database-util/internal/utils"
	"io"
	"log/slog"
	"net/http"
)

type ProductRequest struct {
	Id string `json:"id"`
}

type ProductHandler struct {
	service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) GetProductByUserId(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-user-id")
	if userId == "" {
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing user ID header"})
		return
	}

	products, err := h.service.GetProductsByUserID(r.Context(), userId)
	if err != nil {
		slog.Error("Error Occured", "ERROR", err)
		utils.JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Error in fetching products"})
		return
	}

	utils.JSONResponse(w, http.StatusOK, products)
}

func (h *ProductHandler) StreamProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	err := h.service.StreamProducts(ctx, func(batch []repository.Product) error {
		data, _ := json.Marshal(batch)

		_, err := fmt.Fprintf(w, "data: %s\n\n", data)
		if err != nil {
			return err
		}

		flusher.Flush()
		return nil
	})

	if err != nil {
		slog.Error("stream error", "error", err)
	}

	// End event
	fmt.Fprintf(w, "event: end\ndata: done\n\n")
	flusher.Flush()
}

func (h *ProductHandler) GetProductById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error in parsing body", "Error", err)
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Error in parsing body. Please provide correct body"})
		return
	}

	productRequest, err := utils.ConvertBody[ProductRequest](body)
	if err != nil {
		slog.Error("Error converting request body", "Error", err)
		utils.JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	product, err := h.service.GetProductByID(r.Context(), productRequest.Id)
	if err != nil {
		slog.Error("Error fetching product", "Error", err)
		utils.JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch product"})
		return
	}

	utils.JSONResponse(w, http.StatusOK, product)
}
