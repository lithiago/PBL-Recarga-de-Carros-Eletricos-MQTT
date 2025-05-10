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
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Servidor struct {
	IP     string
	ID     string
	Regiao string
	Client clientemqtt.MQTTClient
	Pontos []*consts.Posto
}

var (
	arquivoPontos = os.Getenv("ARQUIVO_JSON")
)

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

func (s *Servidor) carregarPontos() []*consts.Posto {
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

// Adicionando o método getPostosFromJSON
func (s *Servidor) getPostosFromJSON() ([]*consts.Posto, error) {
	filePath := arquivoPontos
	if filePath == "" {
		return nil, fmt.Errorf("ARQUIVO_JSON não definido")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo JSON: %v", err)
	}

	var mapa map[string][]consts.Posto
	if err := json.Unmarshal(data, &mapa); err != nil {
		return nil, fmt.Errorf("erro ao desserializar o JSON: %v", err)
	}

	estado := os.Getenv("ESTADO")
	postos, ok := mapa[estado]
	if !ok {
		return nil, fmt.Errorf("nenhum dado encontrado para o estado: %s", estado)
	}

	// Converte para []*consts.Posto
	var resultado []*consts.Posto
	for i := range postos {
		resultado = append(resultado, &postos[i])
	}

	return resultado, nil
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

	// Criar uma instância do servidor
	server := &Servidor{
		IP:     ip,
		Client: mqttClient,
	}

	// Registrar o handler para o tópico de reserva
	routerServidor.Register(topics.CarroRequestReserva("+"), func(payload []byte) {
		var msg consts.Mensagem
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Println("Erro ao decodificar mensagem:", err)
			return
		}
		log.Printf("[SERVIDOR] Solicitação de reserva recebida de car/%s", msg.CarroMQTT.ID)
		// Responder ao carro usando a instância do servidor
		server.ResponderCarro(msg.CarroMQTT.ID, "Reservado!") // Alterado para usar 'server'
	})

	// Registrar o handler para o tópico de status
	routerServidor.Register(topics.CarroRequestStatus("+"), func(payload []byte) {
		var msg consts.Mensagem
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Println("Erro ao decodificar mensagem:", err)
			return
		}
		log.Printf("[SERVIDOR] Atualização de status recebida de car/%s", msg.CarroMQTT.ID)
		// Aqui você pode implementar a lógica para lidar com a atualização de status
	})

	return *server
}

func serverCarConnection() {
	server := inicializarServidor()
	server.Pontos = server.carregarPontos()
	server.AssinarEventosDoCarro()
}

func serverAPICommunication(server *Servidor) {
	log.Println("[SERVIDOR] Iniciando comunicação API REST entre servidores com Gin...")

	r := gin.Default()

	r.GET("/postos", func(c *gin.Context) {
		postos, err := server.getPostosFromJSON()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(postos) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Nenhum posto encontrado"})
			return
		}

		// Retorna os postos como JSON
		c.JSON(http.StatusOK, postos)
	})

	// Inicia o servidor HTTP na porta 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("[SERVIDOR] Erro ao iniciar servidor HTTP com Gin: %v", err)
	}
}

func main() {
	log.Println("[SERVIDOR] Inicializando...")

	go serverCarConnection() // Chama a função para iniciar a conexão do carro
	go serverAPICommunication(&Servidor{}) // Passa a instância do servidor para a comunicação API

	select {} // mantém o servidor ativo
}
// mantém o servidor ativo
