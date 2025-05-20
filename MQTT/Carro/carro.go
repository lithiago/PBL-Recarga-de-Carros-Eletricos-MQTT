package main

import (
	consts "MQTT/utils/Constantes"
	topics "MQTT/utils/Topicos"
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	router "MQTT/utils/mqttLib/Router"
	"math/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
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
	userInputChan    = make(chan string)          // Canal para entrada do usuário
	quitChan         = make(chan os.Signal, 1)    // Canal para sinal de encerramento
)

type Carro struct {
	ID                string                 `json:"id"`
	Bateria           float64                `json:"bateria"`
	Clientemqtt       clientemqtt.MQTTClient `json:"-"`
	X                 float64                `json:"x"`
	Y                 float64                `json:"y"`
	CapacidadeBateria float64                `json:"capacidadebateria"`
	Consumobateria    float64                `json:"consumobateria"`
	// Adicionado para a função solicitarRota
}

// Métodos do Carro (permanecem semelhantes, pois publicam diretamente)
func (c *Carro) SolicitarReserva(cidadeDestino string, serverID string) {
	topic := topics.CarroRequestReserva(c.ID, serverID, cidadeDestino)
	log.Println("[CARRO] Publicando solicitação de reserva no tópico: ", topic)
	c.Clientemqtt.Publish(topic, []byte("Aguardando confirmação de reserva!"))
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
	fmt.Print(" 👉 Escolha uma opção: ")
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
			var msgServer consts.Mensagem
			//var paradas map[string][]consts.Parada
			
			// Desserializa a mensagem para o tipo genérico
			msgServer = desserializarMensagem(msg.Payload) 

			

			fmt.Println(">> [Rotas Recebidas]:", msgServer.Msg)
			// for cidade, paradasList := range paradas {
			// 	fmt.Printf("  Cidade: %s\n", cidade)
			// 	for _, parada := range paradasList {
			// 		fmt.Printf("    Posto: %s, ID: %s, X: %.2f, Y: %.2f\n", parada.NomePosto, parada.IDPosto, parada.X, parada.Y)
			// 	}
			// }
			// Adicione lógica para exibir visualmente ou armazenar rotas
		} else {
			log.Printf("[Processador MQTT] Tópico desconhecido ou não tratado especificamente: %s\n", msg.Topic)
		}
	}
	log.Println("[Processador MQTT] Encerrando.")
}

// readUserInput lê a entrada do terminal e envia para o canal userInputChan
func readUserInput(inputChan chan<- string) {
	log.Println("[Entrada Usuário] Iniciado. Digite '3' para sair.")
	for {
		// Não exibir o menu aqui, pois é responsabilidade do loop principal
		// fmt.Print(">> ") // Não é necessário aqui, pois exibirMenu já faz
		var input string
		_, err := fmt.Scanln(&input) // Lê a linha completa
		if err != nil {
			log.Printf("[Entrada Usuário] Erro ao ler entrada: %v\n", err)
			// Pode querer enviar um sinal de erro ou fechar o canal aqui
			return
		}
		inputChan <- strings.TrimSpace(input) // Envia a entrada para o canal
	}
}

func (c *Carro) AssinarRespostaServidor() {
	topicResp := topics.ServerResponseToCar(c.ID)
	c.Clientemqtt.Subscribe(topicResp)
	log.Printf("[CARRO] Subscrito ao tópico: %s\n", topicResp)

	topicRoutes := topics.ServerResponteRoutes(c.ID, "+")
	c.Clientemqtt.Subscribe(topicRoutes)
	log.Printf("[CARRO] Subscrito ao tópico: %s\n", topicRoutes)
}


// handleUserCommand processa os comandos recebidos do canal userInputChan
func (c *Carro) handleUserCommand(command string) {
	switch command {
	case "1": // Solicitar Rota para Destino
		cidades := consts.CidadesArray
		cidadeInicial := consts.CidadeAtualDoCarro(c.X, c.Y)
		var indice int = -1
		for i, _ := range cidades{
			if cidades[i] == strings.ToLower(cidadeInicial){
				indice = i
				break
			}
		}
		if indice != -1 {
			// Remove a cidade atual da lista de opções
			cidades = append(cidades[:indice], cidades[indice+1:]...)
		}
		
		fmt.Println("Cidades disponíveis para rota:")
		for i, cidade := range cidades {
			fmt.Printf("  %d - %s\n", i, cidade)
		}
		fmt.Print("Digite a opção para cidade de destino: ")
		var escolha int
		_, err := fmt.Scanln(&escolha)
		if err != nil || escolha < 0 || escolha >= len(cidades) {
			fmt.Println("Opção inválida.")
			return
		}
		cidadeDestino := cidades[escolha]
		c.solicitarRota(cidadeInicial, cidadeDestino)
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

func main() {
	log.Println("[CARRO] Inicializando aplicação...")

	routerCarro := router.NewRouter()
	mqttClient := *clientemqtt.NewClient(string(consts.Broker), routerCarro)

	// Conectar ao broker MQTT
	conn := mqttClient.Connect()
	if conn.Wait() && conn.Error() != nil {
		log.Fatalf("[CARRO] Erro ao conectar ao broker: %v", conn.Error())
	}
	log.Println("[CARRO] Conectado ao broker MQTT.")

	ip, _ := getLocalIP()
	// Definindo a cidade de origem do carro para o exemplo
	randomX := rand.Float64()*(355.0-60.0) + 60.0
	randomY := rand.Float64()*(270.0-50.0) + 50.0
	carro := Carro{
		ID:                ip,
		Bateria:           60.0,
		Clientemqtt:       mqttClient,
		X:                 randomX,
		Y:                 randomY,
		CapacidadeBateria: 60.0,
		Consumobateria:    0.20,
	}

	// Assinar tópicos necessários no broker MQTT
	carro.AssinarRespostaServidor() // Este método deve apenas subscrever
	// O router.Register é onde você define os handlers para o router,
	// e setupMqttHandlers os configurará para enviar para o canal.
	setupMqttHandlers(routerCarro, carro.ID)

	// Iniciar goroutines de processamento e entrada do usuário
	go processIncomingMqttMessages(&carro) // Goroutine para processar mensagens MQTT do canal
	go readUserInput(userInputChan)         // Goroutine para ler entrada do usuário

	/* // Ação automática: pedir rota após alguns segundos (exemplo)
	go func() {
		time.Sleep(5 * time.Second) // Espera um pouco pra garantir conexão e setup
		log.Println("[CARRO] Ação automática: Solicitando rota para Salvador.")
		carro.solicitarRota(carro.CidadeOrigem, "Salvador")
	}() */

	// Loop principal da aplicação (event loop)
	// Este loop gerencia os eventos da entrada do usuário e o encerramento.
	for {
		carro.exibirMenu() // Exibe o menu antes de cada prompt de entrada
		select {
		case cmd := <-userInputChan: // Recebe comando do usuário
			carro.handleUserCommand(cmd)
			if cmd == "3" { // Se o comando for para sair, o handleUserCommand já enviou para quitChan
				// Não precisa fazer nada aqui, o quitChan abaixo irá pegar
			}
		case <-quitChan: // Recebe sinal de encerramento da aplicação
			fmt.Println("\n[Main] Sinal de encerramento recebido. Iniciando desligamento...")
			 // Salta para o rótulo de encerramento
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


