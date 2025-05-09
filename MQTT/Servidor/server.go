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

func (s *Servidor) adicionarAoArquivo(path string, novoDado interface{}) error {
	// Ler conteúdo atual do JSON
	conteudo, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Decodificar para um mapa com as regiões como chave e os postos como valor
	var mapa map[string][]consts.Posto
	if len(conteudo) > 0 {
		if err := json.Unmarshal(conteudo, &mapa); err != nil {
			return err
		}
	}

	// Recuperar a região (estado) a partir da variável de ambiente
	estado := os.Getenv("ESTADO")
	postos, ok := mapa[estado]
	if !ok {
		return fmt.Errorf("Nenhum dado encontrado para o estado: %s", estado)
	}

	// Encontrar o posto correspondente ao ID do novoDado
	novoPosto := novoDado.(map[string]interface{})
	idNovoPosto := novoPosto["id"].(string)

	// Atualizar o posto ou adicionar um novo
	postoAtualizado := false
	for i, posto := range postos {
		if posto.Id == idNovoPosto {
			// Atualizar as informações do posto
			postos[i].Nome = novoPosto["name"].(string)
			postos[i].X = novoPosto["x"].(float64)
			postos[i].Y = novoPosto["y"].(float64)
			postos[i].Capacidade = novoPosto["capacidade"].(float64)
			postos[i].CustoKW = novoPosto["custoKW"].(float64)
			postoAtualizado = true
			break
		}
	}

	// Se o posto não foi encontrado, adiciona um novo
	if !postoAtualizado {
		novoPostoStruturado := consts.Posto{
			Id:         idNovoPosto,
			Nome:       novoPosto["name"].(string),
			X:          novoPosto["x"].(float64),
			Y:          novoPosto["y"].(float64),
			Capacidade: novoPosto["capacidade"].(float64),
			CustoKW:    novoPosto["custoKW"].(float64),
		}
		postos = append(postos, novoPostoStruturado)
	}

	// Atualizar o mapa com os dados atualizados
	mapa[estado] = postos

	// Reescrever o arquivo com os dados atualizados
	arquivo, err := os.Create(path)
	if err != nil {
		return err
	}
	defer arquivo.Close()

	// Escrever no arquivo com indentação
	encoder := json.NewEncoder(arquivo)
	encoder.SetIndent("", "   ")
	return encoder.Encode(mapa)
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

func serverCarConnection() {
	server := inicializarServidor()
	server.Pontos = server.carregarPontos()
	server.AssinarEventosDoCarro()
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

	go serverCarConnection()
	go serverAPICommunication(&Servidor{})

	// routerServidor := server.Client.Router
	// routerServidor.Register("car/+/request/reservation", func(payload []byte) {
	// 	var msg consts.Mensagem
	// 	if err := json.Unmarshal(payload, &msg); err != nil {
	// 		log.Println("Erro ao decodificar mensagem:", err)
	// 		return
	// 	}
	// 	log.Printf("[SERVIDOR] Solicitação de reserva recebida de car/%s", msg.CarroMQTT.ID)
	// 	server.ResponderCarro(msg.CarroMQTT.ID, "Reservado!")
	// })

	// teste de conexção mqtt
	/*	routerServidor := server.Client.Router

				routerServidor.Register("car/+/request/reservation", func(payload []byte) {
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
				})
				}
		// teste para atualizar o json
			novoPosto := map[string]interface{}{
				"id":         "BA04",            // ID do posto a ser atualizado
				"name":       "Posto 7 - Bahia", // Novo nome do posto
				"x":          777.87,            // Novo valor de X
				"y":          -777.98,           // Novo valor de Y
				"capacidade": 777.0,             // Nova capacidade
				"custoKW":    2.18,              // Novo custo por kWh
			}

			// Chamar a função para atualizar ou adicionar o posto
			err := server.adicionarAoArquivo(arquivo, novoPosto)
			if err != nil {
				// Se ocorrer algum erro, imprime e encerra
				fmt.Println("Erro:", err)
			} else {
				// Caso contrário, confirma que o posto foi atualizado
				fmt.Println("Posto com ID 'BA04' atualizado com sucesso!")
			}
	*/
	select {} // mantém o servidor ativo
}
