package clientemqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttlib "MQTT/mqttLib/Router"
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

func (m *MQTTClient) Connect() {
	token := m.client.Connect()
	token.Wait()
}

func (m *MQTTClient) Subscribe(topic string) {
	m.client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		m.router.Handle(msg.Topic(), msg.Payload())
	})
}

func (m *MQTTClient) Publish(topic string, payload []byte) {
	m.client.Publish(topic, 0, false, payload)
}