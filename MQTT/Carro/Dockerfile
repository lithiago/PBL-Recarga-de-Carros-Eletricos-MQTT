# Etapa de build
FROM golang:1.24 AS builder

WORKDIR /app
COPY ../ .  

# Garante build estático
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -o carro ./Carro/carro.go

# Etapa final
FROM alpine:latest
COPY --from=builder /app/carro /carro
CMD ["/carro"]
