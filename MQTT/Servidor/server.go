package main

import (
	consts "MQTT/utils/Constantes"
	rotaslib "MQTT/utils/Rotas"
	topics "MQTT/utils/Topicos"
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	router "MQTT/utils/mqttLib/Router"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Servidor struct {
	IP     string
	ID     string
	Cidade string
	Client clientemqtt.MQTTClient
	Pontos map[string][]*consts.Posto
}

var (
	arquivoPontos = os.Getenv("ARQUIVO_JSON")
)

// Mapa de cidades para containers e portas
var cidadeConfig = map[string]struct {
	Container string
	Porta     string
}{
	"feiradesantana": {"feiradesantana", "8080"},
	"salvador":       {"salvador", "8082"},
	"ilheus":         {"ilheus", "8081"},
}

// A variavel solicitação é para concatenar a string ao topico evitando multiplas condições
func (s *Servidor) ResponderCarro(carID string, conteudoJSON []byte) {
	topic := topics.ServerResponseToCar(carID)
	log.Printf("[SERVIDOR] Respondendo para: %s", topic)
	s.Client.Publish(topic, conteudoJSON)
}

func (s *Servidor) AssinarEventosDoCarro() {
	topicsToSubscribe := []string{
		topics.CarroRequestReserva("+", s.ID, s.Cidade),
		topics.CarroRequestStatus("+", s.ID, s.Cidade),
		topics.CarroRequestCancel("+", s.ID, s.Cidade),
		topics.CarroRequestRotas("+", s.Cidade),
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

func (s *Servidor) carregarPontos() []consts.Posto {
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
	// Converte para []*consts.Posto
	var resultado []*consts.Posto
	for i := range postos {
		resultado = append(resultado, &postos[i])
	}

	log.Printf("Servidor carregado com %d postos de %s\n", len(resultado), cidade)
	return postos
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
		Cidade: os.Getenv("CIDADE"),
	}
}

func lerRotas() consts.DadosRotas {
	filePath := os.Getenv("ARQUIVO_JSON_ROTAS")
	if filePath == "" {
		panic("ARQUIVO_JSON_ROTAS não definido")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var dados consts.DadosRotas
	if err := json.Unmarshal(data, &dados); err != nil {
		panic(err)
	}
	return dados
}

func desserializarMensagem(payload []byte) consts.Mensagem {
	var msg consts.Mensagem
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Println("Erro ao decodificar mensagem:", err)
		return msg
	}
	return msg
}

// func calcularRotas(rotasPossiveis map[string][]string, trajeto consts.Trajeto) map[int][]string {
//     inicio := trajeto.Inicio
//     destino := trajeto.Destino
//     rotasValidas := make(map[int][]string)

//     if inicio == "" || destino == "" {
//         log.Println("Erro: Início ou destino inválido.")
//         return rotasValidas
//     }

//     log.Printf("Início: %s, Destino: %s", inicio, destino)
//     indiceRotaValida := 0

//     for _, rota := range rotasPossiveis {
//         indiceInicio := -1
//         indiceDestino := -1

//         for i, cidade := range rota {
//             if cidade == inicio && indiceInicio == -1 {
//                 indiceInicio = i
//             }
//             if cidade == destino {
//                 indiceDestino = i
//             }
//         }

//         // Verifica se a sub-rota é válida
//         if indiceInicio != -1 && indiceDestino != -1 && indiceInicio <= indiceDestino {
//             subRota := rota[indiceInicio : indiceDestino+1]
//             rotasValidas[indiceRotaValida] = subRota
//             indiceRotaValida++
//         }
//     }

//     return rotasValidas
// }


func getRotasValidas(rotasPossiveis map[string][]string, trajeto consts.Trajeto) map[string][]string {
	inicio := trajeto.Inicio
	destino := trajeto.Destino
	mapaCompleto := make(map[string][]string)
	// Itera sobre as rotas possíveis
	log.Printf("Início: %s, Destino: %s", inicio, destino)
	for i, rota := range rotasPossiveis {
		var indiceDestino int
		var encontrouDestino bool
		// Percorre as cidades da rota para verificar onde o destino está
		for j, cidade := range rota {

			// Verifica se o destino foi encontrado (case insensitive)
			if strings.EqualFold(cidade, destino) {
				indiceDestino = j
				encontrouDestino = true
				break
			}
		}
		// Se o destino foi encontrado, verifica se o início está na rota antes do destino
		if encontrouDestino {
			var encontrouInicio bool
			// Percorre até o índice do destino para verificar se o início está na rota
			for _, cidade := range rota[:indiceDestino+1] {
				if strings.EqualFold(cidade, inicio) {
					encontrouInicio = true
					break
				}
			}
			// Se o início foi encontrado e está antes do destino, adiciona a rota
			if encontrouInicio {
				mapaCompleto[i] = rota[:indiceDestino+1]
			}
		}
	}

	// Log para verificar o mapa final
	log.Println("Rotas válidas:", mapaCompleto)

	return mapaCompleto
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

func (s *Servidor) ObterPostosDeOutroServidor(url string) ([]*consts.Posto, error) {
	log.Printf("[SERVIDOR] Enviando requisição para %s/postos", url)

	// Cria a requisição HTTP GET
	resp, err := http.Get(url + "/postos/disponiveis")
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição para %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Verifica o status da resposta
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requisição falhou com status %d", resp.StatusCode)
	}

	// Decodifica o JSON da resposta
	var postos []consts.Posto
	if err := json.NewDecoder(resp.Body).Decode(&postos); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	// Convert to []*consts.Posto
	var postosPointers []*consts.Posto
	for i := range postos {
		postosPointers = append(postosPointers, &postos[i])
	}
	log.Printf("[SERVIDOR] Postos recebidos de %s: %+v", url, postos)
	return postosPointers, nil

}

func main() {
	log.Println("[SERVIDOR] Inicializando...")

	server := inicializarServidor()
	server.AssinarEventosDoCarro()

	go serverAPICommunication(&server)
	time.Sleep(10 * time.Second)
	log.Println("[SERVIDOR] Iniciando comunicação MQTT...")

	topic := topics.CarroRequestRotas("+", strings.ToLower(server.Cidade))
	server.Client.Subscribe(topic)
	log.Printf("[SERVIDOR] Assinando tópico: %s", topic)

	routerServidor := server.Client.Router
	routerServidor.Register(topic, func(payload []byte) {
		log.Println("Mensagem Recebida!")
		var conteudoMsg consts.Trajeto
		if err := json.Unmarshal(payload, &conteudoMsg); err != nil {
			log.Println("Erro ao decodificar mensagem:", err)
		}
		dadosRotas := lerRotas()
		rotasValidas := getRotasValidas(dadosRotas.Rotas, conteudoMsg)
		log.Println("Rotas válidas: ", rotasValidas)
		var mapaCompleto = make(map[string][]consts.Posto) // Inicializa o mapa
		var paradas []consts.Parada
		for _, rota := range rotasValidas {
			// AQUI O SERVIDOR DEVE SOLICITAR AOS OUTROS SERVIDORES VIA HTTP OS SEUS PONTOS PARA ASSIM PODER GERAR ROTAS. SÓ FUI OBSERVAR ISSO AGORA. MAS COMO EU VOU MONTAR O MAP PARA PASSAR PRA FUNÇÃO DE GERAR
			// ROTAS?
			for _, cidade := range rota {
				if cidade == server.Cidade {
					mapaCompleto[cidade] = server.carregarPontos() // esse metodo é local
				} else {
					config, exists := cidadeConfig[cidade]
					if !exists {
						log.Printf("Configuração não encontrada para a cidade: %s", cidade)
						continue
					}

					url := "http://servidor-" + config.Container + ":" + config.Porta
					log.Printf("URL: %s", url)
					postos, err := server.ObterPostosDeOutroServidor(url) // obter a partir do http
					if err != nil {
						log.Printf("Erro ao obter postos de outro servidor: %v", err)
						continue
					}
					// Adiciona os postos ao mapa no formato esperado
					var postosSemPonteiro []consts.Posto
					for _, posto := range postos {
						postosSemPonteiro = append(postosSemPonteiro, *posto)
					}
					mapaCompleto[cidade] = postosSemPonteiro
				}
			}

			log.Println("Mapa completo: ", mapaCompleto)
			// paradas := rotaslib.GerarRotas(
			// 	conteudoMsg.CarroMQTT,
			// 	rota,
			// 	dadosRotas.Cidades,
			// 	mapaCompleto,
			// )
			paradas := rotaslib.GerarRotas(conteudoMsg.CarroMQTT, rota, dadosRotas.Cidades, mapaCompleto)
			log.Println("Paradas: ", paradas)

		}
		ConteudoJSON, _ := json.Marshal(paradas)
		topic := topics.ServerResponteRoutes(conteudoMsg.CarroMQTT.ID, server.Cidade)
		log.Printf("[SERVIDOR] Respondendo para: %s", topic)
		server.Client.Publish(topic, ConteudoJSON)
	})

	select {} // mantém o servidor ativo
}