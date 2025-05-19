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
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Servidor struct {
	IP     string
	ID     string
	Cidade string
	Client clientemqtt.MQTTClient
	Pontos map[string][]*consts.Posto
	mu     sync.Mutex
}

type Participante2PC struct {
	PostoID string
	URL     string
}

var (
	arquivoPontos = os.Getenv("ARQUIVO_JSON")
)

// Mapa de cidades para containers e portas
var cidadeConfig = map[string]struct {
	Container string
	Porta     string
}{
	"feiradesantana": {"servidor-feiradesantana", "8080"}, // 172.16.103.10
	"salvador":       {"servidor-salvador", "8082"},       // 172.16.103.9
	"ilheus":         {"servidor-ilheus", "8081"},         // 172.16.103.11
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
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
		var carro struct {
			ID                string  `json:"id"`
			Bateria           float64 `json:"bateria"`
			X                 float64 `json:"x"`
			Y                 float64 `json:"y"`
			Capacidadebateria float64 `json:"capacidadebateria"`
			Consumobateria    float64 `json:"consumobateria"`
		}
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
				p.Fila = append(p.Fila, consts.Carro{
					ID:                carro.ID,
					Bateria:           carro.Bateria,
					X:                 carro.X,
					Y:                 carro.Y,
					Capacidadebateria: carro.Capacidadebateria,
					Consumobateria:    carro.Consumobateria,
				})
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

	r.POST("/2pc/prepare", func(c *gin.Context) {
		var req struct {
			PostoID string       `json:"posto_id"`
			Carro   consts.Carro `json:"carro"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "abort", "error": "Dados inválidos"})
			return
		}
		// Aqui: tente reservar temporariamente a vaga (ex: adicionar na fila como "pendente")
		// Se conseguir reservar:
		c.JSON(http.StatusOK, gin.H{"result": "ok"})
		// Se não conseguir:
		// c.JSON(http.StatusOK, gin.H{"result": "abort"})
	})
	r.POST("/2pc/commit", func(c *gin.Context) {
		var req struct {
			PostoID string       `json:"posto_id"`
			Carro   consts.Carro `json:"carro"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "abort", "error": "Dados inválidos"})
			return
		}
		// Aqui: efetive a reserva (ex: adicione o carro definitivamente à fila)
		c.JSON(http.StatusOK, gin.H{"result": "committed"})
	})

	r.POST("/2pc/abort", func(c *gin.Context) {
		var req struct {
			PostoID string       `json:"posto_id"`
			Carro   consts.Carro `json:"carro"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "abort", "error": "Dados inválidos"})
			return
		}
		// Aqui: desfaça a reserva temporária (ex: remova o carro da fila se estava pendente)
		c.JSON(http.StatusOK, gin.H{"result": "aborted"})
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

// TODO: Testar isso
func (s *Servidor) TwoPhaseCommit(participantes []Participante2PC, carro consts.Carro) error {
	payloadTemplate := `{"posto_id":"%s","carro":%s}`

	// Fase 1: Prepare
	okCount := 0
	for _, p := range participantes {
		carroJSON, _ := json.Marshal(carro)
		payload := fmt.Sprintf(payloadTemplate, p.PostoID, string(carroJSON))
		resp, err := http.Post(p.URL+"/2pc/prepare", "application/json", strings.NewReader(payload))
		if err != nil {
			log.Printf("[2PC] Erro ao enviar prepare para %s: %v", p.URL, err)
			break
		}
		defer resp.Body.Close()
		var res map[string]string
		json.NewDecoder(resp.Body).Decode(&res)
		if res["result"] == "ok" {
			okCount++
		} else {
			break
		}
	}

	// Se todos aceitaram, commit
	if okCount == len(participantes) {
		for _, p := range participantes {
			carroJSON, _ := json.Marshal(carro)
			payload := fmt.Sprintf(payloadTemplate, p.PostoID, string(carroJSON))
			http.Post(p.URL+"/2pc/commit", "application/json", strings.NewReader(payload))
		}
		log.Println("[2PC] Commit enviado para todos os participantes")
		return nil
	}

	// Se algum abortou, abort para todos
	for _, p := range participantes {
		carroJSON, _ := json.Marshal(carro)
		payload := fmt.Sprintf(payloadTemplate, p.PostoID, string(carroJSON))
		http.Post(p.URL+"/2pc/abort", "application/json", strings.NewReader(payload))
	}
	log.Println("[2PC] Abort enviado para todos os participantes")
	return fmt.Errorf("2PC abortado por algum participante")
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

	// Formata o slice de postos como JSON indentado para melhor visualização no log
	postosJSON, err := json.MarshalIndent(postos, "", "  ")
	if err != nil {
		log.Printf("[SERVIDOR] Erro ao formatar postos recebidos de %s: %v", url, err)
	} else {
		log.Printf("[SERVIDOR] Postos recebidos de %s:\n%s", url, string(postosJSON))
	}

	return postosPointers, nil

}

// Adiciona um carro à fila do posto remoto (FUNCIONA)
func (s *Servidor) AdicionarCarroNoPosto(url, postoID string, carro consts.Carro) error {
	endpoint := fmt.Sprintf("%s/postos/%s/adicionar", url, postoID)
	payload, err := json.Marshal(carro)
	if err != nil {
		return fmt.Errorf("erro ao serializar carro: %v", err)
	}

	req, err := http.NewRequest(http.MethodPatch, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição PATCH: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar requisição PATCH: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("requisição PATCH falhou com status %d", resp.StatusCode)
	}

	log.Printf("[SERVIDOR] Carro %s adicionado ao posto %s em %s", carro.ID, postoID, url)
	return nil
}

// Remove um carro da fila do posto remoto (FUNCIONA)
func (s *Servidor) RemoverCarroDoPosto(url, postoID string, carro consts.Carro) error {
	endpoint := fmt.Sprintf("%s/postos/%s/remover", url, postoID)
	payload, err := json.Marshal(carro)
	if err != nil {
		return fmt.Errorf("erro ao serializar carro: %v", err)
	}

	req, err := http.NewRequest(http.MethodPatch, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição PATCH: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar requisição PATCH: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("requisição PATCH falhou com status %d", resp.StatusCode)
	}

	log.Printf("[SERVIDOR] Carro %s removido do posto %s em %s", carro.ID, postoID, url)
	return nil
}

func main() {
	log.Println("[SERVIDOR] Inicializando...")

	server := inicializarServidor()
	// server.AssinarEventosDoCarro()

	go serverAPICommunication(&server)
	time.Sleep(10 * time.Second)
	// log.Println("[SERVIDOR] Iniciando comunicação MQTT...")

	// topic := topics.CarroRequestRotas("+", strings.ToLower(server.Cidade))
	// server.Client.Subscribe(topic)
	// log.Printf("[SERVIDOR] Assinando tópico: %s", topic)

	participantes := []Participante2PC{
		{URL: "http://servidor-feiradesantana:8080", PostoID: "FSA01"},
		{URL: "http://servidor-ilheus:8081", PostoID: "ILH01"},
		{URL: "http://servidor-salvador:8082", PostoID: "SSA01"},
	}

	carro := consts.Carro{
		ID:                "CARRO_TESTE_2PC",
		Bateria:           80.0,
		X:                 100.0,
		Y:                 200.0,
		Capacidadebateria: 100.0,
		Consumobateria:    10.0,
	}

	// O commit chega, mas aparentemente só checa se aquele carro está na fila
	// Preciso mudar para verificar se há qualquer carro na fila, não só o carro que está sendo enviado

	//Teste:
	// Adiciar um carro na fila com a função
	// depois tentar fazer o 2pc
	// deverá falhar
	// depois remover o carro da fila com a função
	// e tentar fazer o 2pc novamente

	log.Println("[TESTE] Iniciando teste do TwoPhaseCommit...")
	err := server.TwoPhaseCommit(participantes, carro)
	if err != nil {
		log.Printf("[TESTE] Falha no 2PC: %v", err)
	} else {
		log.Println("[TESTE] 2PC executado com sucesso!")
	}
	select {} // mantém o servidor ativo
}

//TODO: mandar os postos disponiveis para o carro
//TODO: 2pc
//TODO: tratar concorrência
