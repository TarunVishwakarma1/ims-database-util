# --- Stage 1: Builder ---
FROM golang:1.21-alpine AS builder

# Set the working directory
WORKDIR /app

# Install git (required for fetching Go dependencies)
RUN apk update && apk add --no-cache git

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the statically linked binary. 
# CGO_ENABLED=0 ensures it doesn't rely on host OS C libraries.
# -ldflags="-w -s" strips debugging information to reduce file size.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go-api ./cmd/api/main.go

# --- Stage 2: Final Production Image ---
# 'scratch' is a literally empty image. Zero bloat.
FROM scratch

# We must copy root certificates from the builder so Go can make HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary from the builder stage
COPY --from=builder /go-api /go-api

# Expose the port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/go-api"]