package Constantes

import (
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	mqttlib "MQTT/utils/mqttLib/Router"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Posto struct {
	Id         string
	Nome       string
	Regiao     string
	X          float64
	Y          float64
	Capacidade float64 // Adicionado
	CustoKW    float64 // Adicionado
	Fila       []Carro
}

type MQTTClient struct {
	client mqtt.Client
	router *mqttlib.Router
}

type Carro struct {
	ID             string                 `json:"id"`
	Bateria        float64                `json:"bateria"`
	Clientemqtt    clientemqtt.MQTTClient `json:"-"`
	X              float64                `json:"x"`
	Y              float64                `json:"y"`
	Consumobateria float64                `json:"consumobateria"`
}

type Mensagem struct {
	ConteudoJSON []byte `json:"conteudo"`
	Msg          string `json:"msg"`
}

type Parada struct {
	Cidade       string
	PostoRecarga Posto
}

type Coordenadas struct {
	X float64
	Y float64
}

type DadosRotas struct {
	Cidades map[string]Coordenadas `json:"Cidades"`
	Rotas   map[string][]string    `json:"Rotas"`
}
type Trajeto struct {
	CarroMQTT Carro  `json:"carro"`
	Inicio    string `json:"inicio"`
	Destino   string `json:"destomp"`
}

const Broker = "tcp://mosquitto:1845"
