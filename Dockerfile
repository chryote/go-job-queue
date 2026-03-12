FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o api ./cmd/api/main.go
RUN go build -o worker ./cmd/worker/main.go

# API Stage
FROM alpine:latest AS api
COPY --from=builder /app/api /api
CMD ["/api"]

# Worker Stage
FROM alpine:latest AS worker
COPY --from=builder /app/worker /worker
CMD ["/worker"]