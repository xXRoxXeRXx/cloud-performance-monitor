FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /webhook-logger ./cmd/webhook-logger

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wget
WORKDIR /root/

COPY --from=builder /webhook-logger ./webhook-logger
CMD ["./webhook-logger"]
