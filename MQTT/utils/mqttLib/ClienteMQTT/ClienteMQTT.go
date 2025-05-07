package clientemqtt

import (
	mqttlib "MQTT/utils/mqttLib/Router"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client mqtt.Client
	router *mqttlib.Router
}


func NewClient(broker string, router *mqttlib.Router) *MQTTClient {
	opts := mqtt.NewClientOptions().AddBroker(broker)
	client := mqtt.NewClient(opts)
	return &MQTTClient{client: client, router: router}
}

func (m *MQTTClient) Connect() mqtt.Token{
	return m.client.Connect()
}

func (m *MQTTClient) Subscribe(topic string) {
	m.client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		log.Printf("[MQTT] Recebido no tópico: %s", msg.Topic()) // <--
		m.router.Handle(msg.Topic(), msg.Payload())
	})
}
func (m *MQTTClient) Publish(topic string, payload []byte) {
	m.client.Publish(topic, 0, false, payload)
}