package main

import (
	consts "MQTT/utils/Constantes"
	topics "MQTT/utils/Topicos"
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	router "MQTT/utils/mqttLib/Router"
	"encoding/json"
	"fmt"
	"log"
	//"math/rand"
	"net"
	"time"
)


type Carro struct {
	ID        string                     `json:"id"`
	Bateria   float64                        `json:"bateria"`
	Clientemqtt clientemqtt.MQTTClient `json:"-"`
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Consumobateria float64 `json:"consumobateria"`
}
// tem que ajsutar essa função
func (c *Carro) SolicitarReserva(cidadeDestino string, serverID string) {
	topic := topics.CarroRequestReserva(c.ID, serverID, cidadeDestino)
	log.Println("Topico que o carro publicou: ", topic)
	c.Clientemqtt.Publish(topic, []byte("Aguardando confirmação de reserva!"))
}


func (c *Carro) CancelarReserva(postoID, serverID, cidade string) {
	topic := topics.CarroRequestCancel(c.ID, cidade, serverID)
	c.Clientemqtt.Publish(topic, []byte(postoID))
}

// Função para enviarStatus
/* func (c *Carro) EnviarStatus(serverID) {
	topic := topics.CarroRequestStatus(c.ID)
	// publish payload...
}   */
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


func (c *Carro) publicarAoServidor(conteudoJSON []byte, topico string){
	c.Clientemqtt.Publish(topico, conteudoJSON)
}
func (c *Carro) solicitarRota(cidadeInicial string, cidadeDestino string){
	var trajeto consts.Trajeto
	log.Println("[CARRO] Função solicitarRota foi chamada")
	topic := topics.CarroRequestRotas(c.ID, cidadeDestino)
	log.Printf("[CARRO] Topico: %s", topic)

	trajeto = consts.Trajeto{CarroMQTT: consts.Carro(*c), Inicio: cidadeInicial, Destino: cidadeDestino}
	ConteudoJSON, _ := json.Marshal(trajeto)
	c.publicarAoServidor(ConteudoJSON, topic)
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

func (c *Carro) PorcentagemBateria() float64 {
	return (c.Bateria / 100.0) * 100 // considerando 100kWh como total
}


/* func desserializarMensagem(mensagem []byte) consts.Mensagem{
	var msg consts.Mensagem
		if err := json.Unmarshal(mensagem, &msg); err != nil {
			fmt.Println("Erro ao decodificar:", err)
		}
		return msg
} */
func main() {
	var cidadesPossiveis = []string{"FeiraDeSantana", "Salvador", "Ilheus"}
	log.Println("[CARRO] Inicializando...")

	routerCarro := router.NewRouter()
	mqttClient := *clientemqtt.NewClient(string(consts.Broker), routerCarro)

	conn := mqttClient.Connect()
	if conn.Wait() && conn.Error() != nil {
		log.Fatalf("[CARRO] Erro ao conectar ao broker: %v", conn.Error())
	}

	//rand.Seed(time.Now().UnixNano())
	//num := rand.Intn(len(cidadesPossiveis))
	cidadeOrigem := cidadesPossiveis[2]

	ip, _ := getLocalIP()
	carro := Carro{
		ID:             ip,
		Bateria:        75,
		Clientemqtt:    mqttClient,
		X:              152.0,
		Y:              249.0,
		Consumobateria: 0.15,
	}

	routerCarro.Register(topics.ServerResponseToCar(carro.ID), func(payload []byte) {
		log.Println("[CARRO] [Callback] Resposta direta recebida:")
		fmt.Println("Resposta:", string(payload))
	})

	routerCarro.Register(topics.ServerResponteRoutes(carro.ID, cidadeOrigem), func(payload []byte) {
		log.Println("[CARRO] [Callback] Rotas recebidas:")
		fmt.Println("Rotas:", string(payload))
		var msgServer interface{}

		if err := json.Unmarshal(payload, &msgServer); err != nil {
			fmt.Println("Erro ao decodificar:", err)
		}
	})

	// Inicia escuta MQTT
	//go carro.AssinarResposttaServidor()

	// 🧠 Ação automática: pedir rota para uma cidade aleatória
	go func() {
		time.Sleep(20 * time.Second) // Espera um pouco pra garantir conexão
		destino := "FeiraDeSantana" // ou escolha aleatória diferente da origem
		if cidadeOrigem == "FeiraDeSantana" {
			destino = "Salvador"
		}
		carro.solicitarRota(cidadeOrigem, destino)
	}()

	select {} // mantém o programa vivo
}

