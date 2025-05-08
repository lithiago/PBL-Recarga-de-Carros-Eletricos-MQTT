package main

import (
	consts "MQTT/utils/Constantes"
	topics "MQTT/utils/Topicos"
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	router "MQTT/utils/mqttLib/Router"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

type Servidor struct {
	IP     string
	ID     string
	Regiao string
	Client clientemqtt.MQTTClient
	Pontos []*consts.Posto
}

func (s *Servidor) ResponderCarro(carID string, msg string) {
	topic := topics.ServerResponseToCar(carID)
	log.Printf("[SERVIDOR] Respondendo para: %s", topic)
	s.Client.Publish(topic, []byte(msg))
}

func (s *Servidor) AssinarEventosDoCarro() {
	topicsToSubscribe := []string{
		topics.CarroRequestReserva("+"),
		topics.CarroRequestStatus("+"),
		topics.CarroRequestCancel("+"),
	}
	for _, topic := range topicsToSubscribe {
		s.Client.Subscribe(topic)
	}
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func(s *Servidor) carregarPontos() []*consts.Posto {
	filePath := os.Getenv("ARQUIVO_JSON")
	if filePath == "" {
		panic("ARQUIVO_JSON não definido")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var mapa map[string][]consts.Posto
	if err := json.Unmarshal(data, &mapa); err != nil {
		panic(err)
	}

	estado := os.Getenv("ESTADO")
	postos, ok := mapa[estado]
	if !ok {
		panic(fmt.Sprintf("Nenhum dado encontrado para estado: %s", estado))
	}
	s.Regiao = estado

	// Converte para []*consts.Posto
	var resultado []*consts.Posto
	for i := range postos {
		resultado = append(resultado, &postos[i])
	}

	log.Printf("Servidor carregado com %d postos de %s\n", len(resultado), estado)
	return resultado
}

func inicializarServidor() Servidor {
	routerServidor := router.NewRouter()
	mqttClient := *clientemqtt.NewClient(string(consts.Broker), routerServidor)

	token := mqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		log.Fatalf("Erro ao conectar ao broker: %v", token.Error())
	}

	ip, err := getLocalIP()
	if err != nil {
		log.Printf("Erro ao obter IP local: %v", err)
	}

	return Servidor{
		IP:     ip,
		Client: mqttClient,
	}
}

func main() {
	log.Println("[SERVIDOR] Inicializando...")

	server := inicializarServidor()
	server.Pontos = server.carregarPontos()
	server.AssinarEventosDoCarro()

	//routerServidor := server.Client.Router

	/* routerServidor.Register("car/+/request/reservation", func(payload []byte) {
		var msg consts.Mensagem
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Println("Erro ao decodificar mensagem:", err)
			return
		}
		log.Printf("[SERVIDOR] Solicitação de reserva recebida de car/%s", msg.CarroMQTT.ID)
		server.ResponderCarro(msg.CarroMQTT.ID, "Reservado!")
	})

	routerServidor.Register(topics.CarroRequestCancel("+"), func(payload []byte) {
		var msg consts.Mensagem
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Println("Erro ao decodificar mensagem:", err)
			return
		}
		log.Printf("[SERVIDOR] Solicitação de cancelamento recebida de car/%s", msg.CarroMQTT.ID)
	}) */

	select {} // mantém o servidor ativo
}
