package main

import (
	consts "MQTT/conts"
	clientemqtt "MQTT/mqttLib/ClienteMQTT"
	mqttlib "MQTT/mqttLib/Router"
	"MQTT/topics"
	"encoding/json"
	"fmt"
)


type Servidor struct{
	IP string
	ID string
	Regiao string
	Client  clientemqtt.MQTTClient
}


func (s *Servidor) ResponderCarro(carID string, msg string) {
	topic := topics.ServerResponseToCar(carID)
	s.Client.Publish(topic, []byte(msg))
	// publish...
}
/* 
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

func (s *Servidor) AssinarEventosDoPosto(stationID string) {
	topicsToSubscribe := []string{
		topics.StationStatus(stationID),
		topics.StationEventStarted(stationID),
		topics.StationEventFinished(stationID),
	}
	// subscribe to all topics
}

func (s *Servidor) ObterIdDosCarros () {
	topic := topics.CarroConnect()
	
} */


func main(){
	// VOU CHAVEAR O JSON COM O ENDEREÇO DE IP QUE FOI PRÉ-DEFINIDO

	// Aqui eu implemento um roteador que cuidará do roteamento de topicos para evitar situações de multiplos IFs ou switch Cases.
	router := mqttlib.NewRouter()
	
	// A partir daqui eu vou usar esse router para registrar os topicos e os usar como se listeners (o correto é chamar de callbacks mas fica mais fácil de entender como listener)
	mqttClient := *clientemqtt.NewClient(string(consts.Broker), router)
	server := Servidor{IP:"A", ID:"B", Regiao:"C", Client: mqttClient}
	router.Register("car/+/request/reservation", func(payload []byte) {
		var msg consts.Mensagem
		err := json.Unmarshal(payload, &msg)
		if err != nil {
			fmt.Println("Erro ao decodificar mensagem:", err)
			return
		}
		server.ResponderCarro(msg.CarroMQTT.ID, "Reservado!")
	})

}
