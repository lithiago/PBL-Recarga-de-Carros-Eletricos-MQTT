package main

import (
	consts "MQTT/utils/Constantes"
	topics "MQTT/utils/Topicos"
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	router "MQTT/utils/mqttLib/Router"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Carro struct {
	ID          string                 `json:"id"`
	Bateria     int                    `json:"bateria"`
	X           float64                `json:"x"` // Posição X do carro
	Y           float64                `json:"y"` // Posição Y do carro
	Clientemqtt clientemqtt.MQTTClient `json:"-"` // Cliente MQTT
}

type Servidor struct {
	ID   string  `json:"id"`
	MinX float64 `json:"min_x"`
	MaxX float64 `json:"max_x"`
	MinY float64 `json:"min_y"`
	MaxY float64 `json:"max_y"`
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

func (c *Carro) AssinarRespostaServidor() {
	topic := topics.ServerResponseToCar(c.ID)
	mqttClient := c.Clientemqtt
	mqttClient.Subscribe(topic)
}

func serializarMensagem(msg consts.Mensagem) []byte {
	ConteudoJSON, _ := json.Marshal(msg)
	return ConteudoJSON
}

// Função para carregar servidores a partir de um arquivo JSON
func CarregarServidores() ([]Servidor, error) {
	file, err := os.Open("data/servidores.json") // Ajuste o caminho aqui
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir o arquivo servidores.json: %v", err)
	}
	defer file.Close()

	// Ler o conteúdo do arquivo
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo servidores.json: %v", err)
	}

	// Deserializar o conteúdo JSON para um slice de servidores
	var servidores []Servidor
	if err := json.Unmarshal(content, &servidores); err != nil {
		return nil, fmt.Errorf("erro ao deserializar servidores.json: %v", err)
	}

	return servidores, nil
}

// Função para verificar se o carro está dentro da área de um servidor
func (c *Carro) VerificarServidorProximo(servidores []Servidor) *Servidor {
	for _, servidor := range servidores {
		if c.X >= servidor.MinX && c.X <= servidor.MaxX && c.Y >= servidor.MinY && c.Y <= servidor.MaxY {
			return &servidor
		}
	}
	return nil
}

func main() {
	log.Println("[CARRO] Inicializando...")

	// Carregar lista de servidores do arquivo JSON
	servidores, err := CarregarServidores()
	if err != nil {
		log.Fatalf("[CARRO] Erro ao carregar servidores: %v", err)
	}

	// Definir posição do carro (exemplo fixo, pode ser dinâmica)
	carro := Carro{
		ID:      "001",
		Bateria: 100,
		X:       100.0, // Exemplo de posição X
		Y:       200.0, // Exemplo de posição Y
	}

	// Encontrar o servidor mais próximo
	servidorProximo := carro.VerificarServidorProximo(servidores)
	if servidorProximo == nil {
		log.Fatalf("[CARRO] Nenhum servidor encontrado para a posição atual!")
	}

	log.Printf("[CARRO] Servidor mais próximo: %s", servidorProximo.ID)

	// Conectar ao broker MQTT
	routerCarro := router.NewRouter()
	mqttClient := *clientemqtt.NewClient(string(consts.Broker), routerCarro)

	// Conectar ao broker com verificação
	conn := mqttClient.Connect()
	if conn.Wait() && conn.Error() != nil {
		log.Fatalf("[CARRO] Erro ao conectar ao broker: %v", conn.Error())
	}

	// Associar o cliente MQTT ao carro
	carro.Clientemqtt = mqttClient
	carro.AssinarRespostaServidor()

	// Solicitar reserva ao servidor mais próximo
	carro.SolicitarReserva()

	// Registrar resposta do servidor no tópico
	routerCarro.Register(topics.ServerResponseToCar(carro.ID), func(payload []byte) {
		log.Println("[CARRO] Recebida resposta do servidor")
		fmt.Println("Resposta:", string(payload))
	})

	select {} // Mantém o carro em execução
}
