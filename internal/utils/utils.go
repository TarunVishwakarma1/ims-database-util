package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func ConvertBody[T any](body []byte) (T, error) {
	var data T

	if err := json.Unmarshal(body, &data); err != nil {
		slog.Error("Error in parsing passed body", "error", err)
		return data, err
	}

	return data, nil
}
