FROM golang:1.24

WORKDIR /app

# Copia TUDO (incluindo go.mod, mqttLib, etc.)
COPY .. .

# Entra no diretório do Carro e compila
WORKDIR /app/Carro
RUN go build -o app carro.go

CMD ["./app"]