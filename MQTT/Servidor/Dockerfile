# Etapa de build
FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go build -o servidor ./Servidor/server.go

# Etapa final (imagem enxuta)
FROM alpine:latest

# Copia o binário da etapa anterior
COPY --from=builder /app/servidor /servidor

# Comando de entrada
CMD ["/servidor"]
