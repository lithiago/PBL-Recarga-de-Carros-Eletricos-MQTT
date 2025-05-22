package main

import (
	api "MQTT/Servidor/API"
	consts "MQTT/utils/Constantes"
	rotaslib "MQTT/utils/Rotas"
	topics "MQTT/utils/Topicos"
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	router "MQTT/utils/mqttLib/Router"
	storage "MQTT/utils/storage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Servidor struct {
	IP     string
	ID     string
	Cidade string
	Client clientemqtt.MQTTClient
	Pontos map[string][]*consts.Posto
	carrosConectados map[string]*ConnectedCarStatus
	carrosConectadosMutex sync.Mutex

}

type ConnectedCarStatus struct {
    LastActivity       time.Time
    ReservedPostoID    string // Primeiro posto ou posto local
    CommittedReserva   *consts.Reserva // A reserva completa que foi comitada
    Participantes2PC   []consts.Participante2PC // Os participantes originais do 2PC
}


var (
	arquivoPontos = os.Getenv("ARQUIVO_JSON")
)

// Mapa de cidades para containers e portas
var cidadeConfig = map[string]struct {
	Container string
	Porta     string
}{
	"FSA": {"172.16.201.3", "8080"},
	"ILH": {"171.16.201.16", "8081"},
	"SSA": {"172.16.201.2", "8082"},
}

// A variavel solicitação é para concatenar a string ao topico evitando multiplas condições
func (s *Servidor) ResponderCarro(carID string, conteudoJSON []byte) {
	topic := topics.ServerResponseToCar(carID)
	log.Printf("[SERVIDOR] Respondendo para: %s", topic)
	s.Client.Publish(topic, conteudoJSON)
}

func (s *Servidor) AssinarEventosDoCarro() {
	topicsToSubscribe := []string{
		topics.CarroRequestReserva("+", s.IP, s.Cidade),
		topics.CarroRequestStatus("+", s.IP, s.Cidade),
		topics.CarroRequestCancel("+", s.IP, s.Cidade),
		topics.CarroRequestRotas("+", s.Cidade),
		topics.CarroDesconectado("+"),
	}
	for _, topic := range topicsToSubscribe {
		log.Printf("[SERVIDOR] Assinando tópico: %s", topic)
		s.Client.Subscribe(topic)
	}
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
		return fmt.Errorf("nenhum dado encontrado para o estado: %s", estado)
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
			Id:      idNovoPosto,
			Nome:    novoPosto["name"].(string),
			X:       novoPosto["x"].(float64),
			Y:       novoPosto["y"].(float64),
			CustoKW: novoPosto["custoKW"].(float64),
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
	ip, err := consts.GetLocalIP()
	if err != nil {
		log.Printf("Erro ao obter IP local: %v", err)
	}
	
	routerServidor := router.NewRouter()

	mqttClient := *clientemqtt.NewClient(string(consts.Broker), routerServidor, topics.ServerDesconectado(ip), ip)

	token := mqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		log.Fatalf("Erro ao conectar ao broker: %v", token.Error())
	}

	

	return Servidor{
		IP:     ip,
		Client: mqttClient,
		Cidade: os.Getenv("CIDADE"),
		carrosConectados: make(map[string]*ConnectedCarStatus),
	}
}

func serializarMensagem(msg consts.Mensagem) []byte {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println("Erro ao codificar mensagem:", err)
		return nil
	}
	return msgJSON
}

func (S *Servidor) regitrarHandlersMQTT() {
	routerServidor := S.Client.Router
	routerServidor.Register(topics.CarroRequestReserva("+", S.IP, S.Cidade), func(payload []byte) {
		log.Println("[DEBUG] Carro solicitou reserva")
		var reserva consts.Reserva
		if err := json.Unmarshal(payload, &reserva); err != nil {
			log.Println("Erro ao decodificar mensagem:", err)
			return
		}
		log.Printf("Reserva recebida: %+v\n", reserva)

		// MONTAR URL QUE VAI FAZER PARTE DE PARTICIPANTE2PC EX: "http//:servidor-ip/config.container/portas"
		// Aqui eu tenho que montar um slice dos participantes do 2PC. Cada posto é gerenciado por um servidor especifico.
		// Ex de um PARTICIPANTE2PC a reserva vai passar as cidades que estão presentes nas paradas em postos da rota selecionada
		// Desse modo eu preciso iterar pelo slice de paradas recebido? E então montar as urls e começar a gerar o participante2PC?
		serverURLs := make(map[string]string)
		for cidade, configs := range cidadeConfig {
			// parada agora tem a cidade na struct
			serverURLs[cidade] = fmt.Sprintf("http://%s:%s", configs.Container, configs.Porta)
		}

		var participantes []consts.Participante2PC

		// Itera sobre as paradas da reserva que já contêm as informações necessárias
		for _, parada := range reserva.Paradas {
			if serverURL, ok := serverURLs[parada.Cidade]; ok {
				participantes = append(participantes, consts.Participante2PC{
					URL:     serverURL,
					PostoID: parada.IDPosto,
				})
			} else {
				log.Printf("[ERRO] URL do servidor para a cidade '%s' não encontrada na configuração. Abortando 2PC.\n", parada.Cidade)
				// Enviar uma resposta de erro para o carro aqui.
				return
			}
		}

		// Executa o algoritmo Two-Phase Commit
		topic := topics.ServerReserveStatus(S.IP, reserva.Carro.ID)

		if err := api.TwoPhaseCommit(participantes, reserva.Carro); err != nil {
			log.Printf("[ERRO] Two-Phase Commit falhou: %v\n", err)
			// Lidar com a falha (notificar o carro, etc.)
			msg := consts.Mensagem{
				Conteudo: map[string]interface{}{
					"status": "ERRO",
				},
				Origem: S.Cidade,
				ID:     S.IP,
			}
			S.Client.Publish(topic, serializarMensagem(msg))
		} else {
			log.Println("[INFO] Two-Phase Commit concluído com sucesso!")
			// Lidar com o sucesso (notificar o carro, etc.)

			msg := consts.Mensagem{
				Conteudo: map[string]interface{}{
					"status": "OK",
				},
				Origem: S.Cidade,
				ID:     S.IP,
			}
			S.Client.Publish(topic, serializarMensagem(msg))
			// --- CHAVE: Armazenar o estado da reserva no Coordenador após o COMMIT ---
			S.carrosConectadosMutex.Lock()
			if _, ok := S.carrosConectados[reserva.Carro.ID]; !ok {
				S.carrosConectados[reserva.Carro.ID] = &ConnectedCarStatus{}
			}
			S.carrosConectados[reserva.Carro.ID].CommittedReserva = &reserva       // Armazena a reserva completa
			S.carrosConectados[reserva.Carro.ID].Participantes2PC = participantes // Armazena os participantes
			S.carrosConectados[reserva.Carro.ID].LastActivity = time.Now()         // Atualiza a atividade
			S.carrosConectadosMutex.Unlock()

		}
		log.Println("Publicou no topico: ", topic)

	})
	routerServidor.Register(topics.CarroRequestCancel("+", S.IP, S.Cidade), func(payload []byte) {
		log.Println("[DEBUG] Carro cancelou reserva")

	})
	routerServidor.Register(topics.CarroRequestRotas("+", S.Cidade), func(payload []byte) {
		var conteudoMsg consts.Trajeto
		if err := json.Unmarshal(payload, &conteudoMsg); err != nil {
			log.Println("Erro ao decodificar mensagem:", err)
		}
		dadosRotas := storage.LerRotas()
		rotasValidas := rotaslib.GetRotasValidas(dadosRotas.Rotas, conteudoMsg)
		log.Println("Rotas válidas: ", rotasValidas)
		var mapaCompleto = make(map[string][]consts.Posto) // Inicializa o mapa
		paradas := make(map[string][]consts.Parada)
		for nome, rota := range rotasValidas {
			for _, cidade := range rota {
				if cidade == S.Cidade {
					mapaCompleto[cidade] = storage.CarregarPostos() // esse metodo é local
				} else {
					config, exists := cidadeConfig[cidade]
					if !exists {
						log.Printf("Configuração não encontrada para a cidade: %s", cidade)
						continue
					}
					url := "http://" + config.Container + ":" + config.Porta
					log.Printf("URL: %s", url)
					postos, err := api.ObterPostosDeOutroServidor(url) // obter a partir do http
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

			log.Println("Checando Paradas para a Rota: ", rota)
			paradasArray := rotaslib.GerarRotas(conteudoMsg.CarroMQTT, rota, dadosRotas.Cidades, mapaCompleto)
			if len(paradasArray) != 0 {
				paradas[nome] = paradasArray
			} else {
				log.Printf("⚠️  Rota %s descartada (nenhuma parada válida encontrada).", nome)
			}
			log.Println("Paradas: ", paradas)

		}

		mapInterface := make(map[string]interface{})
		for nome, slice := range paradas {
			mapInterface[nome] = slice
		}

		msg, err := json.Marshal((consts.Mensagem{ID: S.IP, Origem: S.Cidade, Conteudo: mapInterface}))
		if err != nil {
			log.Println("Erro ao codificar mensagem:", err)
			return
		}

		topic := topics.ServerResponteRoutes(conteudoMsg.CarroMQTT.ID, S.Cidade)
		S.Client.Publish(topic, msg)
		log.Println("[DEBUG] JSON final enviado:", string(msg))
	})
	routerServidor.Register(topics.CarroSendsRechargeStart("+", S.IP, S.Cidade), func(payload []byte) {
		log.Println("[DEBUG] Carro informou inicio de recarga")
	})
	routerServidor.Register(topics.CarroSendsRechargeFinish("+", S.IP, S.Cidade), func(payload []byte) {
		log.Println("[DEBUG] Carro informou fim de recarga")
	})
	routerServidor.Register(topics.CarroDesconectado("+"), func(payload []byte){
		S.handleCarroDisconnectedMQTT(payload)
	})

}

func (s *Servidor) handleCarroDisconnectedMQTT(payload []byte) {
	log.Printf("[SERVIDOR] Recebeu mensagem LWT: Payload='%s'\n", string(payload))
	var disconnectedCarPayload map[string]string
	if err := json.Unmarshal(payload, &disconnectedCarPayload); err != nil {
		log.Printf("[SERVIDOR] Erro ao decodificar payload LWT: %v\n", err)
		return
	}
	carroID := disconnectedCarPayload["ID"]
	log.Printf("Carro ID: %s", carroID)
	if carroID != "" {
		log.Printf("[SERVIDOR] Carro %s desconectado inesperadamente. Iniciando processo de limpeza de reservas...\n", carroID)
		s.processCarroDisconnected(carroID) // Chama a lógica de limpeza
	}
}

func (s *Servidor) processCarroDisconnected(carroID string) {
	s.carrosConectadosMutex.Lock()
	carStatus, exists := s.carrosConectados[carroID]
	if !exists {
		s.carrosConectadosMutex.Unlock()
		log.Printf("[SERVIDOR] Carro %s desconectado, mas não tinha reserva ativa registrada por ESTE COORDENADOR.\n", carroID)
		return
	}
	delete(s.carrosConectados, carroID) // Remove do mapa de carros ativos
	s.carrosConectadosMutex.Unlock()        // Libera o lock cedo se a operação remota for demorada

	log.Printf("[SERVIDOR] Carro %s desconectado inesperadamente. Iniciando limpeza de reserva comitada.\n", carroID)

	// Se a reserva foi comitada por ESTE coordenador
	if carStatus.CommittedReserva != nil && len(carStatus.Participantes2PC) > 0 {
		log.Println("Entrou")
		for _, p := range carStatus.Participantes2PC {
			releasePayload := map[string]interface{}{
				"posto_id": p.PostoID,
				"carro":    carStatus.CommittedReserva.Carro,
			}
			releaseJSON, _ := json.Marshal(releasePayload)

			// CORREÇÃO: Enviar requisição HTTP para o endpoint /2pc/release
			resp, err := http.Post(p.URL + "/2pc/release", "application/json", strings.NewReader(string(releaseJSON)))
			if err != nil {
				log.Printf("[SERVIDOR] ERRO: Falha ao enviar requisição de LIBERAÇÃO para %s (posto %s): %v\n", p.URL, p.PostoID, err)
			} else {
				resp.Body.Close()
				log.Printf("[SERVIDOR] Enviado LIBERAÇÃO para %s (posto %s) para carro %s. Resposta: %s\n", p.URL, p.PostoID, carroID, resp.Status)
			}
		}
	} else {
		log.Printf("[SERVIDOR] Carro %s desconectado, mas não tinha reserva multi-servidor comitada ativa neste coordenador.\n", carroID)
	}

	// Lógica para liberar postos LOCAIS que este servidor gerencia (se o carro estava em um deles)
	// Isso é necessário porque o LWT chega a TODOS os servidores assinados no tópico LWT.
	// Cada servidor deve verificar se o carro estava em UM POSTO QUE ELE GERENCIA.
	postosLocais, err := storage.GetPostosFromJSON(os.Getenv("ARQUIVO_JSON")) // Use a variável global
	if err != nil {
		log.Printf("[SERVIDOR] Erro ao ler postos locais para limpeza de %s: %v\n", carroID, err)
		return
	}

	postoLocalAtualizado := false
	for _, p := range postosLocais {
		// Verifica se o carro está na fila deste posto local
		newFila := []consts.Carro{}
		carRemovedFromLocalFila := false
		for _, c := range p.Fila {
			if c.ID != carroID {
				newFila = append(newFila, c)
			} else {
				carRemovedFromLocalFila = true
			}
		}
		if carRemovedFromLocalFila {
			p.Fila = newFila
			postoLocalAtualizado = true
			log.Printf("[SERVIDOR] Carro %s removido da fila do posto LOCAL %s (%s) devido à desconexão.\n", carroID, p.Nome, p.Id)
		}
	}

	if postoLocalAtualizado {
		if err := storage.AtualizarArquivo(os.Getenv("ARQUIVO_JSON"), postosLocais); err != nil {
			log.Printf("[SERVIDOR] Erro ao atualizar posto local após desconexão de %s: %v\n", carroID, err)
		} else {
			log.Printf("[SERVIDOR] Arquivo de postos locais atualizado após limpeza para carro %s.", carroID)
		}
	}
}




func main() {
	log.Println("[SERVIDOR] Inicializando...")
	server := inicializarServidor()
	log.Println("[SERVIDOR] IP:", server.IP)
	server.regitrarHandlersMQTT()
	server.AssinarEventosDoCarro()
	go api.ServerAPICommunication(arquivoPontos)
	time.Sleep(10 * time.Second)
	log.Println("[SERVIDOR] Iniciando comunicação MQTT...")
	select {} // mantém o servidor ativo
}
