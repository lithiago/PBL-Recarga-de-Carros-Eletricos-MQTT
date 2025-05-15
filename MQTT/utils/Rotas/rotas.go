package Rotas

import (
	consts "MQTT/utils/Constantes"
	"log"
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

func GerarRotas(carro consts.Carro, rota []string, cidades map[string]consts.Coordenadas, postosPorCidade map[string][]consts.Posto) map[string][]consts.Parada {
	log.Println("Criando rotas...")
	log.Printf("Bateria: %+v", carro.Bateria)
	log.Printf("Rotas: %v", rota)

	autonomiaTotal := calcularAutonomia(carro.Consumobateria, carro.CapacidadeBateria)
	log.Printf("Autonomia total: %.2f km", autonomiaTotal)

	posicaoCarro := consts.Coordenadas{X: carro.X, Y: carro.Y}
	paradasPorCidade := make(map[string][]consts.Parada)

	for _, cidade := range rota {
		destino := cidades[cidade]

			// Se já estiver na cidade, segue para o próximo destino
			if posicaoCarro.X == destino.X && posicaoCarro.Y == destino.Y {
				log.Printf("O carro já está em %s.", cidade)
				continue
			}
		

		for {
			distanciaAteDestino := calcularDistancia(posicaoCarro, destino)
			log.Printf("Distância até %s: %.2f km", cidade, distanciaAteDestino)

			// Se alcançável, vai direto
			if distanciaAteDestino <= autonomiaTotal {
				log.Printf("Destino %s alcançável sem paradas.", cidade)
				posicaoCarro = destino
				break
			}

			var melhorPosto consts.Posto
			var cidadeRecarga string
			melhorProgresso := distanciaAteDestino
			encontrouPosto := false

			fatorProgresso := 0.8
			distanciaLimite := autonomiaTotal * fatorProgresso

			for cidadeAtual, postos := range postosPorCidade {
				for _, posto := range postos {
					distanciaAtePosto := calcularDistancia(posicaoCarro, consts.Coordenadas{X: posto.X, Y: posto.Y})
					distanciaPostoAteDestino := calcularDistancia(consts.Coordenadas{X: posto.X, Y: posto.Y}, destino)

					if distanciaAtePosto <= autonomiaTotal &&
						distanciaAtePosto < distanciaLimite &&
						distanciaPostoAteDestino < melhorProgresso {

						melhorPosto = posto
						cidadeRecarga = cidadeAtual
						melhorProgresso = distanciaPostoAteDestino
						encontrouPosto = true
					}
				}
			}

			if !encontrouPosto {
				log.Fatalf("Erro: Nenhum posto viável encontrado a partir da posição atual (%+v).", posicaoCarro)
				break
			}

			log.Printf("Parada adicionada: Cidade %s, Posto %+v", cidadeRecarga, melhorPosto)
			parada := consts.Parada{
				Cidade:       cidadeRecarga,
				PostoRecarga: melhorPosto,
			}
			paradasPorCidade[cidadeRecarga] = append(paradasPorCidade[cidadeRecarga], parada)

			// Atualiza a posição do carro para o posto escolhido
			posicaoCarro = consts.Coordenadas{X: melhorPosto.X, Y: melhorPosto.Y}
		}
	}

	log.Println("Paradas formatadas:")
	for cidade, paradas := range paradasPorCidade {
		log.Printf("Cidade: %s", cidade)
		for _, parada := range paradas {
			log.Printf("  Posto: %+v", parada.PostoRecarga)
		}
	}

	return paradasPorCidade
}
