# ==========================================
# STAGE 1: The Builder
# ==========================================
FROM golang:1.25-alpine AS builder

# 1. Install CA certificates 
# (Required if your Go app ever needs to make secure HTTPS requests to other APIs)
RUN apk --no-cache add ca-certificates tzdata

# 2. Set the working directory inside the container
WORKDIR /app

# 3. Cache dependencies
# We copy the mod files first. If they haven't changed, Docker caches the downloaded modules,
# making subsequent builds lightning fast.
COPY go.mod go.sum ./
RUN go mod download

# 4. Copy the rest of the application code
COPY . .

# 5. Compile the Go binary
# - CGO_ENABLED=0: Disables C dependencies, creating a 100% statically linked binary.
# - GOOS=linux GOARCH=amd64: Ensures it compiles for standard Linux servers.
# - ldflags="-w -s": Strips out debugging information and symbol tables to drastically reduce file size.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go-api ./cmd/api/main.go

# ==========================================
# STAGE 2: The Production Image
# ==========================================
# 'scratch' is a special Docker keyword for a completely empty image (0 MB).
FROM scratch

# 1. Copy the timezone data and SSL certificates from the builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 2. Copy our compiled binary from the builder stage
COPY --from=builder /go-api /go-api

# 3. Document the port the container listens on
EXPOSE 8080

# 4. Execute the binary directly (No shell script wrapper)
ENTRYPOINT ["/go-api"]