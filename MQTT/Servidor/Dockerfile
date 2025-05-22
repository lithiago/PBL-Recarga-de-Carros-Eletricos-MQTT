# Etapa de build
FROM golang:1.24 AS builder

WORKDIR /app
COPY ../ . 

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -o servidor ./Servidor/server.go

# Etapa final
FROM alpine:latest
COPY --from=builder /app/servidor /servidor
CMD ["/servidor"]
