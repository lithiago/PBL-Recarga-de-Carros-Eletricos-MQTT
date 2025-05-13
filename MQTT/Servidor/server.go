package main

import (
	consts "MQTT/utils/Constantes"
	//rotaslib "MQTT/utils/Rotas"
	topics "MQTT/utils/Topicos"
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	router "MQTT/utils/mqttLib/Router"
	"encoding/json"
	"fmt"
	"log"
	"net"
	//"net/http"
	"os"
)

type Servidor struct {
	IP     string
	ID     string
	Regiao string
	Client clientemqtt.MQTTClient
	Pontos map[string][]*consts.Posto
}

var (
	arquivo = os.Getenv("ARQUIVO_JSON")
)

// A variavel solicitação é para concatenar a string ao topico evitando multiplas condições
func (s *Servidor) ResponderCarro(carID string, conteudoJSON []byte, solicitacao string) {
	topic := topics.ServerResponseToCar(carID) + solicitacao
	log.Printf("[SERVIDOR] Respondendo para: %s", topic)
	s.Client.Publish(topic, conteudoJSON)
}

func (s *Servidor) AssinarEventosDoCarro() {
	topicsToSubscribe := []string{
		topics.CarroRequestReserva("+", s.ID, s.Regiao),
		topics.CarroRequestStatus("+", s.ID, s.Regiao),
		topics.CarroRequestCancel("+",s.ID, s.Regiao),
		topics.CarroRequestRotas("+", s.Regiao),
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
	// Converte para []*consts.Posto
	var resultado []*consts.Posto
	for i := range postos {
		resultado = append(resultado, &postos[i])
	}

	log.Printf("Servidor carregado com %d postos de %s\n", len(resultado), cidade)
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

func desserializarMensagem(payload []byte) consts.Mensagem{
	var msg consts.Mensagem
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Println("Erro ao decodificar mensagem:", err)
		return msg
	}
	return msg
}


func calcularRotas(rotasPossiveis map[string][]string, trajeto consts.Trajeto) map[int][]string {
	inicio := trajeto.Inicio
	destino := trajeto.Destino
	rotasValidas := make(map[int][]string)

	contador := 0
	for _, rota := range rotasPossiveis {
		indiceInicio := -1
		indiceDestino := -1

		for i, cidade := range rota {
			if cidade == inicio && indiceInicio == -1 {
				indiceInicio = i
			}
			if cidade == destino {
				indiceDestino = i
			}
		}

		if indiceInicio != -1 && indiceDestino != -1 && indiceInicio < indiceDestino {
			subRota := rota[indiceInicio : indiceDestino+1]
			rotasValidas[contador] = subRota
			contador++
		}
	}

	return rotasValidas
}

/* func buscarPostosDeCidade(cidade string) []*consts.Posto {
    url := fmt.Sprintf("http://%s:8080/postos?cidade=%s", servidorDestino, cidade) // MONTAR A URL
    resp, err := http.Get(url)
    // tratar erro, parsear o JSON, etc...
} */

func main() {
	log.Println("[SERVIDOR] Inicializando...")
	
	server := inicializarServidor()
	server.AssinarEventosDoCarro()
	server.Regiao = os.Getenv("CIDADE")
	
	log.Printf("[SERVIDOR] Região: %s", server.Regiao)
	topic := topics.CarroRequestRotas("+", server.Regiao)
	log.Printf("[SERVIDOR] Topico: %s", topic)

	server.Client.Subscribe(topic)
	routerServidor := server.Client.Router
	routerServidor.Register(topic, func(payload []byte) {
		log.Println("Mensagem Recebida!")
		/* var conteudoMsg consts.Trajeto
		msg := desserializarMensagem(payload)
		if err := json.Unmarshal(msg.ConteudoJSON, &conteudoMsg); err != nil {
			log.Println("Erro ao decodificar mensagem:", err)
		}		
		dadosRotas := lerRotas()
		rotasValidas := calcularRotas(dadosRotas.Rotas, conteudoMsg)
		var mapaCompleto map[string][]*consts.Posto
		for _, rota := range rotasValidas {
			log.Printf("Calculando rota %d: %v\n", i+1, rota)
			
			// AQUI O SERVIDOR DEVE SOLICITAR AOS OUTROS SERVIDORES VIA HTTP OS SEUS PONTOS PARA ASSIM PODER GERAR ROTAS. SÓ FUI OBSERVAR ISSO AGORA. MAS COMO EU VOU MONTAR O MAP PARA PASSAR PRA FUNÇÃO DE GERAR 
			// ROTAS? 
			for _, cidade := range rota{
				if cidade == server.Regiao{
					mapaCompleto[cidade] = server.carregarPontos() // esse metodo é local
				} else {
					mapaCompleto[cidade] = buscarPontosDeCidade(cidade) // obter a partir do http
				}
			}
			paradas := rotaslib.GerarRotas(
				conteudoMsg.CarroMQTT,
				rota,
				dadosRotas.Cidades,
				mapaCompleto,
			)
			
		}
		ConteudoJSON, _ := json.Marshal(msg)
		server.ResponderCarro(conteudoMsg.CarroMQTT.ID, ConteudoJSON, "/routes") */
	})

	// teste de conexção mqtt
	/*
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
