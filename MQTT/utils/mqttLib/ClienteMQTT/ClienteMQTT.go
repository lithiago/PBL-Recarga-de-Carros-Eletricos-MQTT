package clientemqtt

import (
	mqttlib "MQTT/utils/mqttLib/Router"
	"encoding/json"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	Client mqtt.Client
	Router *mqttlib.Router
}



func NewClient(broker string, router *mqttlib.Router, LWTtopic string, ID string) *MQTTClient {
	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetCleanSession(true)
	lwtPayload := map[string]string{
		"ID": ID,
		"Motivo": "Desconexão inesperada",
	}
	lwtJSON, err := json.Marshal(lwtPayload)
	if err !=nil{
		log.Fatalf("Erro ao serializar LWT Payload.")
	}
	opts.SetWill(LWTtopic, string(lwtJSON), 1, false)
	client := mqtt.NewClient(opts)
	return &MQTTClient{Client: client, Router: router}
}

func (m *MQTTClient) Connect() mqtt.Token{
	return m.Client.Connect()
}

func (m *MQTTClient) Subscribe(topic string) {
	m.Client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		log.Printf("[MQTT] Recebido no tópico: %s", msg.Topic()) // <--
		m.Router.Handle(msg.Topic(), msg.Payload())
	})
}
func (m *MQTTClient) Publish(topic string, payload []byte) {
	token := m.Client.Publish(topic, 0, false, payload)
	if token.Error() != nil {
        log.Printf("Erro ao publicar: %v", token.Error())
    } else {
        log.Println("Mensagem publicada com sucesso")
    }
}