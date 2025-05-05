FROM golang:1.24

WORKDIR /app

# Copia TUDO (incluindo go.mod, mqttLib, etc.)
COPY .. .

# Entra no diretório do Servidor e compila
WORKDIR /app/Servidor
RUN go build -o app server.go

CMD ["./app"]