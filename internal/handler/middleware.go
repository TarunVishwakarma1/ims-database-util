package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"
)

func RequireHMAC(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timestampStr := r.Header.Get("x-timestamp")
			clientSignature := r.Header.Get("x-signature")

			if timestampStr == "" || clientSignature == "" {
				http.Error(w, "Missing security headers", http.StatusUnauthorized)
				return
			}

			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				http.Error(w, "Invalid timestamp", http.StatusUnauthorized)
				return
			}
			requestTime := time.UnixMilli(timestamp)
			if time.Since(requestTime) > 10*time.Second || time.Until(requestTime) > 10*time.Second {
				http.Error(w, "Request expired", http.StatusUnauthorized)
				return
			}

			payload := r.Method + ":" + r.URL.Path + ":" + timestampStr

			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write([]byte(payload))
			expectedSignature := hex.EncodeToString(mac.Sum(nil))

			if !hmac.Equal([]byte(clientSignature), []byte(expectedSignature)) {
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
