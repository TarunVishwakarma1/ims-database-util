package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"
)

// RequireHMAC returns an HTTP middleware constructor that enforces HMAC-SHA256 request authentication using the provided secret.
// 
// The middleware requires requests to include `x-timestamp` (milliseconds since epoch) and `x-signature` headers.
// It rejects requests when either header is missing, when `x-timestamp` cannot be parsed as a base-10 int64, or when the timestamp
// is more than 10 seconds in the past or more than 10 seconds in the future (responding with 401 and an explanatory message).
// The signing payload is `METHOD:PATH:timestamp` (e.g. `GET:/health:1610000000000`). The middleware computes an HMAC-SHA256 of that
// payload using `secret`, hex-encodes the digest, and compares it to `x-signature` using a constant-time comparison.
// If the signature matches the computed value the request is forwarded to the wrapped handler; otherwise the middleware responds
// with 401 Unauthorized and "Invalid signature".
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
