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
	log.Printf("[CARRO] Publicando no tópico: %s", topic)
	c.Clientemqtt.Publish(topic, serializarMensagem(msg))
}


/* func (c *Carro) CancelarReserva() {
	topic := topics.CarroRequestCancel(c.ID)
	// publish payload...
}

func (c *Carro) EnviarStatus() {
	topic := topics.CarroRequestStatus(c.ID)
	// publish payload...
}  */
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

/* func desserializarMensagem(mensagem []byte) consts.Mensagem{
	var msg consts.Mensagem
		if err := json.Unmarshal(mensagem, &msg); err != nil {
			fmt.Println("Erro ao decodificar:", err)
		}
		return msg
} */
func main(){
	log.Println("[CARRO] Inicializando...")
	routerCarro := router.NewRouter()

	mqttClient := *clientemqtt.NewClient(string(consts.Broker), routerCarro)

	// Conectar ao broker com verificação
	conn := mqttClient.Connect()
	if conn.Wait() && conn.Error() != nil {
		log.Fatalf("[CARRO] Erro ao conectar ao broker: %v", conn.Error())
	}

	carro := Carro{ID: "001", Bateria: 100, Clientemqtt: mqttClient}
	carro.AssinarRespostaServidor()
	carro.SolicitarReserva()

	routerCarro.Register(topics.ServerResponseToCar(carro.ID), func(payload []byte) {
		log.Println("[CARRO] Recebida resposta do servidor")

		fmt.Println("Resposta:", string(payload))
	})


	select {} // Mantém o carro em execução

}
