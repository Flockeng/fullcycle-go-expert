# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copia arquivos de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia código fonte
COPY . .

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia o binário compilado
COPY --from=builder /app/bin/server .

# Expõe a porta 8080
EXPOSE 8080

# Comando para executar
CMD ["./server"]
