# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
# CGO_ENABLED=0 for static linking, GOOS=linux for cross-compilation if needed
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o taskmaster-app .

# Stage 2: Create the final small image
FROM alpine:latest

WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/taskmaster-app .

# Expose port (the port the app listens on)
EXPOSE 8080

# Command to run the executable
CMD ["./taskmaster-app"]