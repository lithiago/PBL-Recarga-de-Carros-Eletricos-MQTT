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

type Servidor struct {
	IP     string
	ID     string
	Regiao string
	Client clientemqtt.MQTTClient
	Pontos map[string]*consts.Posto
}

func (s *Servidor) ResponderCarro(carID string, msg string) {
	topic := topics.ServerResponseToCar(carID)
	log.Printf("[SERVIDOR] Respondendo para: %s", topic)
	s.Client.Publish(topic, []byte(msg))
}

func (s *Servidor) NotificarCarro(carID string) {
	topic := topics.ServerNotifyCar(s.ID, carID)
	// publish...
}

func (s *Servidor) ComandoReservarPosto(stationID string) {
	topic := topics.ServerCommandReserve(stationID)
	// publish...
}

func (s *Servidor) ComandoCancelarReserva(stationID string) {
	topic := topics.ServerCommandCancel(stationID)
	// publish...
}

func (s *Servidor) AssinarEventosDoCarro() {
	topicsToSubscribe := []string{
		topics.CarroRequestReserva("+"),
		topics.CarroRequestStatus("+"),
		topics.CarroRequestCancel("+"),
	}

	for _, topic := range topicsToSubscribe{
		s.Client.Subscribe(topic)
	}
}



func main() {
	log.Println("[SERVIDOR] Inicializando...")

	routerServidor := router.NewRouter()
	mqttClient := *clientemqtt.NewClient(string(consts.Broker), routerServidor)

	// Conectar ao broker com verificação
	conn := mqttClient.Connect()
	if conn.Wait() && conn.Error() != nil {
		log.Fatalf("[SERVIDOR] Erro ao conectar ao broker: %v", conn.Error())
	}

	server := Servidor{IP: "A", ID: "B", Regiao: "C", Client: mqttClient}

	// Registrar handler com suporte a '+'
	mqttClient.Subscribe("car/+/request/reservation")
	routerServidor.Register("car/+/request/reservation", func(payload []byte) {
		log.Println("[SERVIDOR] Mensagem recebida em car/+/request/reservation")
		var msg consts.Mensagem
		if err := json.Unmarshal(payload, &msg); err != nil {
			fmt.Println("Erro ao decodificar:", err)
			return
		}
		server.ResponderCarro(msg.CarroMQTT.ID, "Reservado!")
	})


	select {} // Mantém o servidor em execução
}
