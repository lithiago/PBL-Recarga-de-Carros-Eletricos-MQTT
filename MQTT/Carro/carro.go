package main

import (
	consts "MQTT/utils/Constantes"
	topics "MQTT/utils/Topicos"
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	router "MQTT/utils/mqttLib/Router"
	"encoding/json"
	"fmt"
	"log"
)

type Carro struct {
	ID       string                     `json:"id"`
	Bateria  int                        `json:"bateria"`
	Clientemqtt clientemqtt.MQTTClient `json:"-"`
}

func (c *Carro) SolicitarReserva() {
	topic := topics.CarroRequestReserva(c.ID)

	msg := consts.Mensagem{
		CarroMQTT: consts.Carro{
			ID:      c.ID,
			Bateria: c.Bateria,
		},
		Msg: "Carro solicitando reserva!",
	}

	c.Clientemqtt.Publish(topic, serializarMensagem(msg))
}

/* 
func (c *Carro) CancelarReserva() {
	topic := topics.CarroRequestCancel(c.ID)
	// publish payload...
}

func (c *Carro) EnviarStatus() {
	topic := topics.CarroRequestStatus(c.ID)
	// publish payload...
}
 */ 
func (c *Carro) AssinarRespostaServidor() {
	topic := topics.ServerResponseToCar(c.ID)
	mqttClient := c.Clientemqtt
	mqttClient.Subscribe(topic)
	// subscribe...
}

func serializarMensagem(msg consts.Mensagem) []byte{
	ConteudoJSON, _ := json.Marshal(msg)
	return ConteudoJSON
}
func main(){
	log.Println("Funcionando.")
	routerCarro := router.NewRouter()
	mqttClient := *clientemqtt.NewClient(string(consts.Broker), routerCarro)
	carro := Carro{ID: 	"001", Bateria: 100, Clientemqtt: mqttClient}
	// Registrar handlers para os topicos em que o carro irá assinar
	routerCarro.Register(topics.ServerResponseToCar(carro.ID), func(payload []byte) {
		fmt.Println("Resposta do Servidor:", string(payload))
	})
	carro.SolicitarReserva()

}
