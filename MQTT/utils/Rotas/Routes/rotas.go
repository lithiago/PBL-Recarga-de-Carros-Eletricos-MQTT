package routes

import (
	consts "MQTT/utils/Constantes"
	"math"
)

// Calcula a distância euclidiana entre dois pontos
func calcularDistancia(destino, origem consts.Coordenadas) float64 {
	return math.Sqrt(math.Pow(destino.X-origem.X, 2) + math.Pow(destino.Y-origem.Y, 2))
}

// Calcula quantos quilômetros o carro pode andar com a bateria fornecida
func calcularAutonomia(consumoKW, capacidadeBateria float64) float64 {
	if consumoKW == 0 {
		return 0 // Evita divisão por zero
	}
	return capacidadeBateria / consumoKW
}

// Gera paradas para uma rota específica (lista de cidades)
func gerarRotas(carro consts.Carro, rota []string, cidades map[string]consts.Coordenadas, postosPorCidade map[string][]consts.Posto) map[string][]consts.Parada {
	paradasPorCidade := make(map[string][]consts.Parada)
	autonomiaTotal := calcularAutonomia(carro.Consumobateria, 100)
	autonomiaAtual := calcularAutonomia(carro.Consumobateria, carro.Bateria)
	posicaoAtual := consts.Coordenadas{X: carro.X, Y: carro.Y}

	for _, cidade := range rota {
		destino, existe := cidades[cidade]
		if !existe {
			continue
		}

		distanciaTotal := calcularDistancia(destino, posicaoAtual)
		numeroDeParadas := distanciaTotal / autonomiaTotal
		if numeroDeParadas < 1 && calcularDistancia(posicaoAtual, destino) <= autonomiaAtual {
			// Consegue chegar direto
			posicaoAtual = destino
			continue
		}

		distanciaPorParada := distanciaTotal / math.Ceil(numeroDeParadas)
		fatorProgresso := 0.8

		for {
			menorDist := math.MaxFloat64
			var postoMaisProximo consts.Posto
			encontrou := false

			for cidadeAtual, postos := range postosPorCidade {
				for _, posto := range postos {
					dist := calcularDistancia(consts.Coordenadas{X: posto.X, Y: posto.Y}, posicaoAtual)

					if dist <= autonomiaAtual && dist >= fatorProgresso*distanciaPorParada && dist < menorDist {
						menorDist = dist
						postoMaisProximo = posto
						cidade = cidadeAtual
						encontrou = true
					}
				}
			}

			if !encontrou {
				break
			}

			parada := consts.Parada{
				Cidade:       cidade,
				PostoRecarga: postoMaisProximo,
			}
			paradasPorCidade[cidade] = append(paradasPorCidade[cidade], parada)
			posicaoAtual = consts.Coordenadas{X: postoMaisProximo.X, Y: postoMaisProximo.Y}
			autonomiaAtual = autonomiaTotal

			if calcularDistancia(destino, posicaoAtual) <= autonomiaAtual {
				posicaoAtual = destino
				break
			}
		}
	}

	return paradasPorCidade
}





