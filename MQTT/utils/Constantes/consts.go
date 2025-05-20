package Constantes

import (
	clientemqtt "MQTT/utils/mqttLib/ClienteMQTT"
	mqttlib "MQTT/utils/mqttLib/Router"
	"encoding/json"
	"math"

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
	Conteudo json.RawMessage `json:"conteudo"` 
	Msg          string `json:"msg"`
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
	X float64
	Y float64
}

// Calcula a distância euclidiana entre dois pontos
func CalcularDistancia(destino, origem Coordenadas) float64 {
	return math.Sqrt(math.Pow(destino.X-origem.X, 2) + math.Pow(destino.Y-origem.Y, 2))
}

const Broker = "tcp://mosquitto:1845"
