
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN ls -lR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /agent ./cmd/agent

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and curl for health checks
RUN apk --no-cache add ca-certificates curl tzdata

WORKDIR /
COPY --from=builder /agent /agent

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health/live || exit 1

EXPOSE 8080
CMD ["/agent"]
