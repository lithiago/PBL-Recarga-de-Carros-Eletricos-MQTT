package main

import (
	consts "MQTT/utils/Constantes"
	topics "MQTT/utils/Topicos"
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	router "MQTT/utils/mqttLib/Router"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings" // Para lowercasing do comando do usuário
	"time"
)

// MqttMessage representa uma mensagem MQTT recebida para ser enviada pelo canal
type MqttMessage struct {
	Topic   string
	Payload []byte
}

// Canais globais para comunicação entre goroutines
var (
	incomingMqttChan = make(chan MqttMessage, 100) // Canal para mensagens MQTT recebidas, com buffer
	userInputChan    = make(chan string)           // Canal para entrada do usuário
	quitChan         = make(chan os.Signal, 1)     // Canal para sinal de encerramento
	promptChan       = make(chan Prompt)
)

type Prompt struct {
	Pergunta   string
	RespostaCh chan string
}

type Carro struct {
	ID                string                 `json:"id"`
	Bateria           float64                `json:"bateria"`
	Clientemqtt       clientemqtt.MQTTClient `json:"-"`
	X                 float64                `json:"x"`
	Y                 float64                `json:"y"`
	CapacidadeBateria float64                `json:"capacidadebateria"`
	Consumobateria    float64                `json:"consumobateria"`
	CidadeAtual       string                 `json:"cidadeatual"`
	// Adicionado para a função solicitarRota
}

func (c *Carro) SolicitarReserva(rotas map[string][]consts.Parada, cidadeDestino string, serverID string) {

	rotasIndexadas := []string{}
	for nome, paradas := range rotas {
		fmt.Printf("\n[%d] %s:\n", len(rotasIndexadas), nome)
		for i, parada := range paradas {
			fmt.Printf("  \t [%d] %s (ID: %s)\n", i+1, parada.NomePosto, parada.IDPosto)
			fmt.Printf("      \t Localização: (X: %.2f, Y: %.2f)\n", parada.X, parada.Y)
		}
		rotasIndexadas = append(rotasIndexadas, nome)
	}

	input := perguntarUsuario("Digite o número da rota desejada: ")
	escolha, _ := strconv.Atoi(input)

	if escolha < 0 || escolha >= len(rotasIndexadas) {
		fmt.Println("❌ Escolha inválida.")
		return

	}

	nomeRotaEscolhida := rotasIndexadas[escolha]
	fmt.Printf("Você escolheu a rota: %s\n", nomeRotaEscolhida)
	paradasEscolhidas := rotas[nomeRotaEscolhida]
	// Enviar a rota escolhida para o servidor

	reserva := consts.Reserva{
		Carro: consts.Carro{
			ID:                c.ID,
			Bateria:           c.Bateria,
			X:                 c.X,
			Y:                 c.Y,
			CapacidadeBateria: c.CapacidadeBateria,
			Consumobateria:    c.Consumobateria,
		},
		Paradas: paradasEscolhidas,
	}

	ConteudoJSON, err := json.Marshal(reserva)
	if err != nil {
		log.Printf("[ERRO] Falha ao serializar mensagem de reserva: %v\n", err)
		return
	}

	topic := topics.CarroRequestReserva(c.ID, serverID, cidadeDestino)
	log.Println("[CARRO] Publicando solicitação de reserva no tópico: ", topic)
	c.publicarAoServidor(ConteudoJSON, topic)
}

func (c *Carro) CancelarReserva(postoID, serverID, cidade string) {
	topic := topics.CarroRequestCancel(c.ID, cidade, serverID)
	log.Println("[CARRO] Publicando cancelamento de reserva no tópico: ", topic)
	c.Clientemqtt.Publish(topic, []byte(postoID))
}

func serializarMensagem(msg consts.Mensagem) []byte {
	ConteudoJSON, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[ERRO] Falha ao serializar mensagem: %v\n", err)
		return nil
	}
	return ConteudoJSON
}

func (c *Carro) publicarAoServidor(conteudoJSON []byte, topico string) {
	if conteudoJSON == nil {
		log.Println("[CARRO] Não foi possível publicar: conteúdo JSON é nulo.")
		return
	}
	log.Printf("[CARRO] Publicando no tópico: %s com payload: %s\n", topico, string(conteudoJSON))
	c.Clientemqtt.Publish(topico, conteudoJSON)
}

func (c *Carro) solicitarRota(cidadeInicial string, cidadeDestino string) {
	log.Println("[CARRO] Função solicitarRota foi chamada")
	topic := topics.CarroRequestRotas(c.ID, cidadeDestino)
	log.Printf("[CARRO] Topico para solicitação de rota: %s", topic)

	trajeto := consts.Trajeto{
		CarroMQTT: consts.Carro{
			ID:                c.ID,
			Bateria:           c.Bateria,
			X:                 c.X,
			Y:                 c.Y,
			CapacidadeBateria: c.CapacidadeBateria,
			Consumobateria:    c.Consumobateria,
		},
		Inicio:  cidadeInicial,
		Destino: cidadeDestino,
	}
	ConteudoJSON, err := json.Marshal(trajeto)
	if err != nil {
		log.Printf("[ERRO] Falha ao serializar trajeto para rota: %v\n", err)
		return
	}
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
	return (c.Bateria / c.CapacidadeBateria) * 100
}

func desserializarMensagem(mensagem []byte) consts.Mensagem {
	var msg consts.Mensagem
	if err := json.Unmarshal(mensagem, &msg); err != nil {
		fmt.Printf("[ERRO] Erro ao decodificar mensagem: %v\n", err)
		// Retorna uma MsgServer vazia ou com erro sinalizado
		return consts.Mensagem{}
	}
	return msg
}

func (c *Carro) exibirMenu() {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("          🚀 MENU PRINCIPAL 🚀        ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  🆔 Carro ID: %s \n", c.ID)
	fmt.Printf("  🔋 Bateria: %.2f%%\n", c.PorcentagemBateria())
	fmt.Println("  1️⃣  | Solicitar Rota para Destino")
	fmt.Println("  2️⃣  | Simular Viagem") // Exemplo de nova opção
	fmt.Println("  3️⃣  | Encerrar Conexão")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func setupMqttHandlers(router *router.Router, carID string) {
	// Handler para respostas diretas do servidor ao carro
	router.Register(topics.ServerResponseToCar(carID), func(payload []byte) {
		log.Println("[CARRO] [Callback MQTT] Resposta direta recebida. Enviando para canal...")
		incomingMqttChan <- MqttMessage{
			Topic:   topics.ServerResponseToCar(carID),
			Payload: payload,
		}
	})

	// Handler para rotas/paradas do servidor
	router.Register(topics.ServerResponteRoutes(carID, "+"), func(payload []byte) {
		log.Println("[CARRO] [Callback MQTT] Rotas/Paradas recebidas. Enviando para canal...")
		incomingMqttChan <- MqttMessage{
			Topic:   topics.ServerResponteRoutes(carID, "+"), // O '+' deve ser substituído pelo tópico real se necessário para processamento
			Payload: payload,
		}
	})

	router.Register(topics.ServerReserveStatus("+", carID), func(payload []byte) {
		log.Println("[CARRO] [Callback MQTT] recebido status de reserva do servido. Envinado para canal...")
		incomingMqttChan <- MqttMessage{
			Topic:   topics.ServerReserveStatus("+", carID),
			Payload: payload,
		}
	})
	// Adicione outros handlers conforme necessário
}

// Esta função tá com pouca legibilidade, vou organizar ela depois
func processIncomingMqttMessages(car *Carro) {
	log.Println("[Processador MQTT] Iniciado.")
	for msg := range incomingMqttChan { // O loop termina quando o canal é fechado
		log.Printf("[Processador MQTT] Recebeu mensagem do tópico: %s\n", msg.Topic)

		// Lógica para diferenciar e processar mensagens baseada no tópico
		// Você pode usar funções de utilidade do seu pacote 'topics' para isso
		if strings.HasPrefix(msg.Topic, topics.ServerResponseToCar(car.ID)) {
			fmt.Printf(">> [Resposta Servidor] %s\n", string(msg.Payload))
			// Lógica específica para respostas diretas (ex: confirmações)
		} else if strings.HasPrefix(msg.Topic, topics.ServerResponteRoutes(car.ID, "")) { // Prefixo para rotas
			//var paradas map[string][]consts.Parada

			// Desserializa a mensagem para o tipo genérico
			msgServer := desserializarMensagem(msg.Payload)

			fmt.Println(">> Rotas Recebidas do IP :", msgServer.ID)
			paradasMap := make(map[string][]consts.Parada)

			for nome, v := range msgServer.Conteudo {
				// Primeiro, transforma o slice genérico em JSON
				bytes, err := json.Marshal(v)
				if err != nil {
					log.Println("Erro ao serializar trecho:", err)
					continue
				}

				// Depois desserializa como []Parada
				var paradas []consts.Parada
				err = json.Unmarshal(bytes, &paradas)
				if err != nil {
					log.Println("Erro ao converter para []Parada:", err)
					continue
				}

				paradasMap[nome] = paradas
			}

			car.SolicitarReserva(paradasMap, msgServer.Origem, msgServer.ID)

			//Adicione lógica para exibir visualmente ou armazenar rotas
		} else if strings.HasPrefix(msg.Topic, topics.ServerReserveStatus("+", car.ID)) {
			msgServer := desserializarMensagem(msg.Payload)
			fmt.Println("Status de Reserva recebido do IP:", msgServer.ID)
			reserveStatus := make(map[string]string)
			for k, v := range msgServer.Conteudo{
				bytes, _ := json.Marshal(v)
				var status string
				err := json.Unmarshal(bytes, &status)
				if err != nil{
					log.Println("Erro ao converter para string")
				}
				reserveStatus[k] =  status
			}
			if reserveStatus["status"] == "OK"{
				log.Println("Reserva bem sucedida")
				
			} else if reserveStatus["status"] == "ERRO"{
				log.Println("Erro ao reserver postos.")
				log.Println("[SOLICITE OUTRA ROTA]")
				log.Println("Cidade destino: ", msgServer.Origem)
				car.solicitarRota(car.CidadeAtual, msgServer.Origem)

			}

		} else {
			log.Printf("[Processador MQTT] Tópico desconhecido ou não tratado especificamente: %s\n", msg.Topic)
		}
	}
	log.Println("[Processador MQTT] Encerrando.")
}

func (c *Carro) AssinarRespostaServidor() {
	topicResp := topics.ServerResponseToCar(c.ID)
	c.Clientemqtt.Subscribe(topicResp)
	log.Printf("[CARRO] Subscrito ao tópico: %s\n", topicResp)

	topicRoutes := topics.ServerResponteRoutes(c.ID, "+")
	c.Clientemqtt.Subscribe(topicRoutes)
	log.Printf("[CARRO] Subscrito ao tópico: %s\n", topicRoutes)

	topicReserveStatus := topics.ServerReserveStatus("+", c.ID)
	c.Clientemqtt.Subscribe(topicReserveStatus)
	log.Printf("[CARRO] Subscrito ao tópico: %s\n", topicReserveStatus)
}

func (c *Carro) selecionarCidade() string {
	// Remove a cidade atual da lista de cidades disponíveis
	cidades := make([]string, 0, len(consts.CidadesArray))
	for _, cidade := range consts.CidadesArray {
		log.Printf("[CARRO] Cidade: %s\n [CARRO] Cidade Atual: %s\n", cidade, c.CidadeAtual)
		if cidade != c.CidadeAtual {
			cidades = append(cidades, cidade)
		}
	}

	fmt.Println("Cidades disponíveis para rota:")
	for i, cidade := range cidades {
		fmt.Printf("  %d - %s\n", i, cidade)
	}

	input := perguntarUsuario(strings.TrimSpace("Digite a opção para cidade de destino: "))
	escolha, _ := strconv.Atoi(input)
	if escolha < 0 || escolha >= len(cidades) {
		fmt.Println("Opção inválida. Tente novamente.")
		return ""
	}
	cidadeDestino := cidades[escolha]
	return cidadeDestino
}

// handleUserCommand processa os comandos recebidos do canal userInputChan
func (c *Carro) handleUserCommand(command string) {
	switch command {
	case "1": // Solicitar Rota para Destino
		// Remove a cidade atual da lista de cidades disponíveis
		cidades := make([]string, 0, len(consts.CidadesArray))
		for _, cidade := range consts.CidadesArray {
			log.Printf("[CARRO] Cidade: %s\n [CARRO] Cidade Atual: %s\n", cidade, c.CidadeAtual)
			if cidade != c.CidadeAtual {
				cidades = append(cidades, cidade)
			}
		}

		fmt.Println("Cidades disponíveis para rota:")
		for i, cidade := range cidades {
			fmt.Printf("  %d - %s\n", i, cidade)
		}
		fmt.Print("Digite a opção para cidade de destino: ")
		var escolha int
		fmt.Scanln(&escolha)
		if escolha < 0 || escolha >= len(cidades) {
			fmt.Println("Opção inválida. Tente novamente.")
			return
		}
		cidadeDestino := cidades[escolha]
		c.solicitarRota(c.CidadeAtual, cidadeDestino)
	case "2": // Simular Viagem (Exemplo de nova opção)
		fmt.Println("Pensar em algo para colocar aqui")
		// Aqui você poderia iniciar uma goroutine para simular o movimento do carro, consumo de bateria, etc.
	case "3": // Encerrar Conexão
		fmt.Println("Precisa implementar encerramento de conexões")
		quitChan <- os.Interrupt // Envia um sinal para o canal de encerramento
	default:
		fmt.Println("Opção inválida. Tente novamente.")
	}
}

func readUserInput() {
	log.Println("[Entrada Usuário] Iniciado.")
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case prompt := <-promptChan:
			fmt.Print(prompt.Pergunta)
			input, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("[Entrada Usuário] Erro ao ler entrada: %v\n", err)
				prompt.RespostaCh <- ""
			} else {
				prompt.RespostaCh <- strings.TrimSpace(input)
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func perguntarUsuario(pergunta string) string {
	respCh := make(chan string)
	promptChan <- Prompt{
		Pergunta:   pergunta,
		RespostaCh: respCh,
	}
	return <-respCh
}

func main() {
	log.Println("[CARRO] Inicializando aplicação...")
	ip, _ := getLocalIP()

	routerCarro := router.NewRouter()
	mqttClient := *clientemqtt.NewClient(string(consts.Broker), routerCarro, topics.CarroDesconectado(ip), ip)

	// Conectar ao broker MQTT
	conn := mqttClient.Connect()
	if conn.Wait() && conn.Error() != nil {
		log.Fatalf("[CARRO] Erro ao conectar ao broker: %v", conn.Error())
	}
	log.Println("[CARRO] Conectado ao broker MQTT.")

	// Definindo a cidade de origem do carro para o exemplo
	randomX := rand.Float64()*(355.0-60.0) + 60.0
	randomY := rand.Float64()*(270.0-50.0) + 50.0
	cidadeInicial := consts.CidadeAtualDoCarro(randomX, randomY)
	log.Printf("Cidade [%s]: (%2f, %2f) \n", cidadeInicial, randomX, randomY)
	carro := Carro{
		ID:                ip,
		Bateria:           60.0,
		Clientemqtt:       mqttClient,
		X:                 randomX,
		Y:                 randomY,
		CapacidadeBateria: 60.0,
		Consumobateria:    0.20,
		CidadeAtual:       cidadeInicial,
	}

	// Assinar tópicos necessários no broker MQTT
	carro.AssinarRespostaServidor()
	// e setupMqttHandlers os configurará para enviar para o canal.
	setupMqttHandlers(routerCarro, carro.ID)

	// Iniciar goroutines de processamento e entrada do usuário
	go processIncomingMqttMessages(&carro) // Goroutine para processar mensagens MQTT do canal
	go readUserInput()                     // Goroutine para ler entrada do usuário


	
	for {

		carro.exibirMenu() // Exibe o menu antes de cada prompt de entrada
		opcao := strings.TrimSpace(perguntarUsuario("Digite a opção desejada: "))
		switch opcao {
		case "1":
			cidadeDestino := carro.selecionarCidade()
			carro.solicitarRota(carro.CidadeAtual, cidadeDestino)
		case "2":
			log.Println("Fazer alguma coisa")
		case "3":
			log.Println("Desconectado")
			break // Adiciona a quebra do loop
		default:
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}

	// A PARTIR DAQUI OS TRECHOS SÃO DE DESLIGAMENTO MAS AINDA NÃO ESTÃO IMPLEMENTADOS. MAS A IDEIA É QUE SEJA ASSIM

	// Processo de desligamento gracioso
	fmt.Println("[Main] Fechando canais e desconectando do broker...")

	// Fecha o canal de mensagens MQTT para sinalizar à goroutine processIncomingMqttMessages para parar
	close(incomingMqttChan)
	// Dá um pequeno tempo para a goroutine de processamento terminar de consumir as mensagens restantes
	time.Sleep(time.Second)

	/* // Desconecta do broker MQTT
	if mqttClient.IsConnected() { // Supondo que você tenha um método IsConnected()
		mqttClient.Disconnect(250)
		fmt.Println("[Main] Desconectado do broker MQTT.")
	} else {
		fmt.Println("[Main] Cliente MQTT já estava desconectado.")
	} */

	fmt.Println("[Main] Aplicação encerrada com sucesso.")
}
