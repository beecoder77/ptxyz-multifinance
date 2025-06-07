# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies and migrate tool
RUN apk --no-cache add ca-certificates tzdata curl bash postgresql-client && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    curl -L https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh > /usr/local/bin/wait-for-it.sh && \
    chmod +x /usr/local/bin/wait-for-it.sh

# Copy binary and configs from builder
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations
COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# Set timezone
ENV TZ=Asia/Jakarta

# Expose port
EXPOSE 8080

# Run the application with migrations
ENTRYPOINT ["./docker-entrypoint.sh"] 