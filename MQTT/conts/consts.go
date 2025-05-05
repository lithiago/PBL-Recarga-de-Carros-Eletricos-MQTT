package consts

import (
	clientemqtt "MQTT/mqttLib/ClienteMQTT"
	mqttlib "MQTT/mqttLib/Router"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Posto struct{
	Id string
}


type MQTTClient struct {
	client mqtt.Client
	router *mqttlib.Router
}

type Carro struct {
	ID       string                     `json:"id"`
	Bateria  int                        `json:"bateria"`
	Clientemqtt clientemqtt.MQTTClient `json:"-"`
}

type Mensagem struct {
	CarroMQTT Carro `json:"carro"`
	Msg       string `json:"msg"`
}

const Broker = "tcp://mosquitto:1883"

