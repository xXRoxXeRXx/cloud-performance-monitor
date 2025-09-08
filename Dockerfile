
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /agent ./cmd/agent

# Runtime stage
FROM alpine:latest
WORKDIR /
COPY --from=builder /agent /agent
EXPOSE 8080
CMD ["/agent"]
