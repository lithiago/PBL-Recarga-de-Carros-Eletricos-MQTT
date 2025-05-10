package Constantes

import (
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	mqttlib "MQTT/utils/mqttLib/Router"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Posto struct {
	Id       string
	Nome     string
	Regiao   string
	X        float64
	Y        float64
	Capacidade float64 // Adicionado
	CustoKW  float64   // Adicionado
	Fila     []Carro
}

type MQTTClient struct {
	client mqtt.Client
	router *mqttlib.Router
}

type Carro struct {
	ID        string                     `json:"id"`
	Bateria   int                        `json:"bateria"`
	Clientemqtt clientemqtt.MQTTClient `json:"-"`
}

type Mensagem struct {
	CarroMQTT Carro `json:"carro"`
	Msg       string `json:"msg"`
}

const Broker = "tcp://mosquitto:1883"
