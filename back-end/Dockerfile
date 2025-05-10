# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary dependencies
RUN apk --no-cache add ca-certificates gcc musl-dev git

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download && go mod tidy && go mod verify

# Copy source code
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o tracepost-larvae

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy built binary from builder stage
COPY --from=builder /app/tracepost-larvae .

# Copy .env file
COPY .env .

# Expose port
EXPOSE 8080

# Command to run the application
CMD ["./tracepost-larvae"]