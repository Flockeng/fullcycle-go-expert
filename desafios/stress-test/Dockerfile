# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /stresstester

# Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /stresstester /stresstester

ENTRYPOINT ["/stresstester"]
