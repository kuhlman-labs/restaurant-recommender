# ========================
# 1. Build Stage
# ========================
FROM golang:1.24 AS builder

# Create and set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to take advantage of Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of your source code
COPY . .

# Build the Go app
# For a web API, ensure it listens on a configurable port (e.g., 8080)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# ========================
# 2. Runtime Stage
# ========================
FROM debian:bullseye-slim AS runtime

# Update and install CA certificates
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    update-ca-certificates --fresh && \
    rm -rf /var/lib/apt/lists/*

# Create and set working directory inside the container
WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/main /app/main

# (Optional) If you have static assets, config files, etc., copy them as well
# COPY --from=builder /app/static /app/static

# Expose the port your application listens on
# By default, Azure App Service for Containers listens on port 80, 
# but you can instruct the container to listen on 8080 (then configure App Settings).
EXPOSE 80

# Set the entrypoint or command to run Go binary
CMD ["/app/main"]