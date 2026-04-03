package handler

import (
	"encoding/json"
	"fmt"
	"ims-database-util/internal/repository"
	"ims-database-util/internal/service"
	"log/slog"
	"net/http"
)

type CustomerHandler struct {
	service service.CustomerService
}

func NewCustomerHandler(service service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) StreamCustomers(w http.ResponseWriter, r *http.Request) {
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

	err := h.service.StreamCustomers(ctx, func(batch []repository.Customer) error {
		data, err := json.Marshal(batch)
		if err != nil {
			return fmt.Errorf("failed to marshal batch: %w", err)
		}

		_, err = fmt.Fprintf(w, "data: %s\n\n", data)
		if err != nil {
			return err
		}

		flusher.Flush()
		return nil
	})

	if err != nil {
		slog.Error("stream error", "error", err)
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	fmt.Fprintf(w, "event: end\ndata: done\n\n")
	flusher.Flush()
}
