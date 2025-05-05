# Etapa de build
FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go build -o carro ./Carro/carro.go

# Etapa final
FROM alpine:latest

COPY --from=builder /app/carro /carro

CMD ["/carro"]