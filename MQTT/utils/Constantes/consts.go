package Constantes

import (
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	mqttlib "MQTT/utils/mqttLib/Router"
	"encoding/json"
	"log"
	"math"
	"net"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Posto struct {
	Id      string  `json:"id"`
	Nome    string  `json:"nome"`
	Cidade  string  `json:"cidade"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	CustoKW float64 `json:"custokw"` // Adicionado
	Fila    []Carro `json:"fila"`
}

type MQTTClient struct {
	client mqtt.Client
	router *mqttlib.Router
}

type Carro struct {
	ID                string                 `json:"id"`
	Bateria           float64                `json:"bateria"`
	Clientemqtt       clientemqtt.MQTTClient `json:"-"`
	X                 float64                `json:"x"`
	Y                 float64                `json:"y"`
	CapacidadeBateria float64                `json:"capacidadebateria"`
	Consumobateria    float64                `json:"consumobateria"`
}

type Mensagem struct {
	Conteudo map[string]interface{} `json:"conteudo"`
	Origem string				 `json:"origem"`
	ID          string `json:"msg"`
}

type MsgServer struct {
	ID       string `json:"id"`
	Cidade   string `json:"cidade"`
	Conteudo json.RawMessage `json:"conteudo"`
}

type Parada struct {
	NomePosto string  `json:"nomeposto"`
	IDPosto   string  `json:"idposto"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
}

type DadosRotas struct {
	Cidades map[string]Coordenadas `json:"cidades"`
	Rotas   map[string][]string    `json:"rotas"`
}
type Trajeto struct {
	CarroMQTT Carro  `json:"carro"`
	Inicio    string `json:"inicio"`
	Destino   string `json:"destino"`
}

type Coordenadas struct {
	Nome string
	X float64
	Y float64
}

type Reserva struct{
	CarroID string `json:"carroid"`
	Paradas []Parada `json:"paradas"`
}




var cidades = map[string]struct {
	x, y, raio float64
}{
	"FSA": {x: 97.67, y: 200.0, raio: 225.9},  // Raio aproximado calculado
	"SSA":         {x: 152.0, y: 249.0, raio: 210.1},
	"ILH":           {x: 300.67, y: 101.0, raio: 225.9},
}


var CidadesArray = []string{"FSA", "SSA", "ILH"}


// Calcula a distância euclidiana entre dois pontos
func CalcularDistancia(destino, origem Coordenadas) float64 {
	return math.Sqrt(math.Pow(destino.X-origem.X, 2) + math.Pow(destino.Y-origem.Y, 2))
}

// Função para calcular a distância entre dois pontos
func distancia(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

// Função para determinar em qual cidade o carro está
func CidadeAtualDoCarro(xCarro, yCarro float64) string {
	for cidade, info := range cidades {
		d := distancia(xCarro, yCarro, info.x, info.y)
		if d <= info.raio {
			log.Println("Carro está em:", cidade)
			return cidade
		}
	}
	return "Fora de cobertura"
}

func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

const Broker = "tcp://mosquitto:1845"
