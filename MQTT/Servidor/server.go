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
	"time"

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

	cidade := os.Getenv("CIDADE")
	postos, ok := mapa[cidade]
	if !ok {
		panic(fmt.Sprintf("Nenhum dado encontrado para cidade: %s", cidade))
	}
	s.Regiao = cidade

	// Converte para []*consts.Posto
	var resultado []*consts.Posto
	for i := range postos {
		resultado = append(resultado, &postos[i])
	}

	log.Printf("Servidor carregado com %d postos de %s\n", len(resultado), cidade)
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

func (s *Servidor) atualizarArquivo(filePath string, postos []*consts.Posto) error {
	// Converte os postos para o formato de mapa esperado no JSON
	cidade := os.Getenv("CIDADE")
	if cidade == "" {
		return fmt.Errorf("CIDADE não definida")
	}

	mapa := map[string][]consts.Posto{
		cidade: make([]consts.Posto, len(postos)),
	}

	for i, posto := range postos {
		mapa[cidade][i] = *posto
	}

	// Serializa o mapa para JSON
	data, err := json.MarshalIndent(mapa, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar os dados para JSON: %v", err)
	}

	// Escreve os dados no arquivo
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("erro ao escrever no arquivo JSON: %v", err)
	}

	return nil
}

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

	cidade := os.Getenv("CIDADE")
	postos, ok := mapa[cidade]
	if !ok {
		return nil, fmt.Errorf("nenhum dado encontrado para a cidade: %s", cidade)
	}

	// Converte para []*consts.Posto
	var resultado []*consts.Posto
	for i := range postos {
		resultado = append(resultado, &postos[i])
	}

	return resultado, nil
}

// retorna os postos disponíveis (sem nenhum carro na fila))
func (s *Servidor) getPostosDisponiveis() ([]*consts.Posto, error) {
	postos, err := s.getPostosFromJSON()
	if err != nil {
		return nil, err
	}

	var postosDisponiveis []*consts.Posto
	for _, posto := range postos {
		if len(posto.Fila) == 0 { // Verifica se a fila está vazia
			postosDisponiveis = append(postosDisponiveis, posto)
		}
	}

	if len(postosDisponiveis) == 0 {
		return nil, fmt.Errorf("nenhum posto disponível encontrado")
	}

	return postosDisponiveis, nil
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

	r.GET("/postos/disponiveis", func(c *gin.Context) {
		postos, err := server.getPostosDisponiveis()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(postos) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "nenhum posto disponível encontrado"})
			return
		}
		c.JSON(http.StatusOK, postos)
	})

	r.PATCH("/postos/:id/adicionar", func(c *gin.Context) {
		id := c.Param("id")
		var carro consts.Carro
		if err := c.ShouldBindJSON(&carro); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
			return
		}

		postos, err := server.getPostosFromJSON()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var postoAtualizado *consts.Posto
		for _, p := range postos {
			if p.Id == id {
				// Verifica se já existe um carro na fila
				if len(p.Fila) > 0 {
					c.JSON(http.StatusConflict, gin.H{"error": "Já existe um carro na fila"})
					return
				}
				// Adiciona o carro à fila
				p.Fila = append(p.Fila, carro)
				postoAtualizado = p
				break
			}
		}

		if postoAtualizado == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Posto não encontrado"})
			return
		}

		if err := server.atualizarArquivo(arquivoPontos, postos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o arquivo JSON"})
			return
		}

		c.JSON(http.StatusOK, postoAtualizado)
	})

	r.PATCH("/postos/:id/remover", func(c *gin.Context) {
		id := c.Param("id")
		var carro consts.Carro
		if err := c.ShouldBindJSON(&carro); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
			return
		}

		postos, err := server.getPostosFromJSON()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var postoAtualizado *consts.Posto
		for _, p := range postos {
			if p.Id == id {
				for i, c := range p.Fila {
					if c.ID == carro.ID {
						p.Fila = append(p.Fila[:i], p.Fila[i+1:]...)
						postoAtualizado = p
						break
					}
				}
				break
			}
		}

		if postoAtualizado == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Posto ou carro não encontrado"})
			return
		}

		if err := server.atualizarArquivo(arquivoPontos, postos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o arquivo JSON"})
			return
		}

		c.JSON(http.StatusOK, postoAtualizado)
	})

	// Inicia o servidor HTTP na porta 8080
	porta := os.Getenv("PORTA")
	if porta == "" {
		log.Fatalf("[SERVIDOR] Erro ao iniciar servidor HTTP com Gin: variável de ambiente PORTA não definida")
	}
	r.Run(":" + porta)
}

// TODO : IMPLEMENTAR O 2PC
// TODO: adicionar o mutex no server

// funciona
func (s *Servidor) ObterPostosDeOutroServidor(url string) ([]*consts.Posto, error) {
	log.Printf("[SERVIDOR] Enviando requisição para %s/postos", url)

	// Cria a requisição HTTP GET
	resp, err := http.Get(url + "/postos")
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição para %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Verifica o status da resposta
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requisição falhou com status %d", resp.StatusCode)
	}

	// Decodifica o JSON da resposta
	var postos []*consts.Posto
	if err := json.NewDecoder(resp.Body).Decode(&postos); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	log.Printf("[SERVIDOR] Postos recebidos de %s: %+v", url, postos)
	return postos, nil
}

// func exemploUso(server *Servidor) {
// 	urlOutroServidor := "http://localhost:8080" // URL do outro servidor

// 	postos, err := server.ObterPostosDeOutroServidor(urlOutroServidor)
// 	if err != nil {
// 		log.Printf("Erro ao obter postos do servidor %s: %v", urlOutroServidor, err)
// 		return
// 	}

// 	log.Printf("Postos recebidos do servidor %s: %+v", urlOutroServidor, postos)
// }

func main() {
	log.Println("[SERVIDOR] Inicializando...")

	server := inicializarServidor()
	server.Pontos = server.carregarPontos()
	server.AssinarEventosDoCarro()

	go serverAPICommunication(&server) //antes criava um novo servidor, agora usa o mesmo
	time.Sleep(60 * time.Second)
	urlOutroServidor := "http://localhost:8080"
	log.Printf("Tentando obter postos do servidor %s...", urlOutroServidor)
	postos, err := server.ObterPostosDeOutroServidor(urlOutroServidor)
	if err != nil {
		log.Printf("Erro ao obter postos do servidor %s: %v", urlOutroServidor, err)
	} else {
		postosJSON, _ := json.MarshalIndent(postos, "", "  ")
		log.Printf("Postos recebidos do servidor %s: %+v", urlOutroServidor, string(postosJSON))
	}

	select {} // mantém o servidor ativo
}
