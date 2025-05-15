package Constantes

import (
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	mqttlib "MQTT/utils/mqttLib/Router"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Posto struct {
	Id      string  `json:"id"`
	Nome    string  `json:"nome"`
	Cidade  string  `json:"cidade"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	CustoKW float64 `json:"custokw"` // Adicionado
	Fila    []Carro `json:"fila"`
}

type MQTTClient struct {
	client mqtt.Client
	router *mqttlib.Router
}

type Carro struct {
	ID                string                 `json:"id"`
	Bateria           float64                `json:"bateria"`
	Clientemqtt       clientemqtt.MQTTClient `json:"-"`
	X                 float64                `json:"x"`
	Y                 float64                `json:"y"`
	Capacidadebateria float64                `json:"capacidadebateria"`
	Consumobateria    float64                `json:"consumobateria"`
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
	Cidades map[string]Coordenadas `json:"cidades"`
	Rotas   map[string][]string    `json:"rotas"`
}
type Trajeto struct {
	CarroMQTT Carro  `json:"carro"`
	Inicio    string `json:"inicio"`
	Destino   string `json:"destino"`
}

const Broker = "tcp://mosquitto:1845"
