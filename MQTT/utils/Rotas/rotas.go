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


func GerarRotas(carro consts.Carro, rota []string, cidades map[string]consts.Coordenadas, postosPorCidade map[string][]consts.Posto) {
    log.Println("Criando rotas...")
    log.Printf("Bateria: %+v", carro.Bateria)
    log.Printf("Rotas: %v", rota)

    // Calcular Autonomia do carro
    autonomiaTotal := calcularAutonomia(carro.Consumobateria, carro.CapacidadeBateria)
    log.Printf("Autonomia total: %.2f km", autonomiaTotal)
    posicaoCarro := consts.Coordenadas{X: carro.X, Y: carro.Y}
    paradasPorCidade := make(map[string][]consts.Parada)

    for _, cidade := range rota {
        destino := cidades[cidade]

        // Verifica se o carro já está na cidade de destino
        if posicaoCarro.X == destino.X && posicaoCarro.Y == destino.Y {
            log.Printf("O carro já está em %s. Seguindo para o próximo destino.", cidade)
            continue
        }

        // Calcula a distância até o destino
        distanciaTotal := calcularDistancia(destino, posicaoCarro)
        log.Printf("Distância até %s: %.2f km", cidade, distanciaTotal)

        if distanciaTotal <= autonomiaTotal {
            log.Printf("Destino %s alcançável sem paradas. Distância: %.2f", cidade, distanciaTotal)
            posicaoCarro = destino
            continue
        }

        // Iniciar cálculo de paradas
        distanciaPorParada := distanciaTotal / autonomiaTotal
        fatorProgresso := 0.8
        for {
            var postoMaisProximo consts.Posto
            var cidadeRecarga string
            distanciaX := 0.0

            // Encontre o posto mais próximo no caminho
            for cidadeAtual, postos := range postosPorCidade {
                for _, posto := range postos {
                    distanciaAtePosto := calcularDistancia(consts.Coordenadas{X: posto.X, Y: posto.Y}, posicaoCarro)
                    if distanciaAtePosto > autonomiaTotal {
                        log.Printf("Posto %s está fora da autonomia atual do carro.", posto.Nome)
                        continue
                    } else if distanciaAtePosto > distanciaX && distanciaAtePosto < fatorProgresso*distanciaPorParada {
                        distanciaX = distanciaAtePosto
                        postoMaisProximo = posto
                        cidadeRecarga = cidadeAtual
                    }
                }
            }

            // Adiciona a parada ao mapa
            log.Printf("Parada adicionada: Cidade %s, Posto %+v", cidadeRecarga, postoMaisProximo)
            parada := consts.Parada{
                Cidade:       cidadeRecarga,
                PostoRecarga: postoMaisProximo,
            }
            paradasPorCidade[cidadeRecarga] = append(paradasPorCidade[cidadeRecarga], parada)
            posicaoCarro = consts.Coordenadas{X: postoMaisProximo.X, Y: postoMaisProximo.Y}

            // Verifica se o carro pode chegar ao destino após a parada
            if calcularDistancia(destino, posicaoCarro) <= autonomiaTotal {
                log.Printf("Destino %s alcançado após a parada.", cidade)
                posicaoCarro = destino
                break
            }
        }
    }
}


/*
func GerarRotas(carro consts.Carro, rota []string, cidades map[string]consts.Coordenadas, postosPorCidade map[string][]consts.Posto) map[string][]consts.Parada {
	log.Println("Gerando rotas...")
	log.Printf("Bateria: %+v", carro)
	log.Printf("Rotas: %v", rota)
	paradasPorCidade := make(map[string][]consts.Parada)

	// Calcular a autonomia do carro com base no consumo
	autonomiaTotal := calcularAutonomia(carro.Consumobateria, 100)
	if autonomiaTotal == 0 {
		log.Println("Erro: autonomia total é zero. Verifique os dados do carro.")
		return paradasPorCidade
	}

	// Autonomia atual com a carga da bateria
	autonomiaAtual := calcularAutonomia(carro.Consumobateria, carro.Bateria)
	log.Printf("Autonomia total: %.2f km", autonomiaTotal)
	posicaoAtual := consts.Coordenadas{X: carro.X, Y: carro.Y}

	// Se o carro já está em uma cidade da rota, não precisa calcular a distância de origem
	for _, cidade := range rota {
		// Pega as coordenadas da cidade de destino
		destino := cidades[cidade]

		// Verifica se a cidade de origem é a mesma do destino
		if posicaoAtual.X == destino.X && posicaoAtual.Y == destino.Y {
			log.Printf("O carro já está em %s. Seguindo para o próximo destino.", cidade)
			continue
		}

		// Calcula a distância até a cidade de destino
		distanciaTotal := calcularDistancia(destino, posicaoAtual)
		log.Printf("Distância até %s: %.2f km", cidade, distanciaTotal)

		// Calcula o número de paradas necessárias
		numeroDeParadas := distanciaTotal / autonomiaTotal
		log.Printf("Número de paradas necessárias: %.2f", numeroDeParadas)

		// Se o destino está dentro da autonomia, não precisa de paradas
		if numeroDeParadas < 1 && distanciaTotal <= autonomiaAtual {
			log.Printf("Destino %s alcançável sem paradas. Distância: %.2f", cidade, distanciaTotal)
			posicaoAtual = destino
			continue
		}

		// Se precisa de paradas, calcula a distância entre cada uma delas
		distanciaPorParada := distanciaTotal / math.Ceil(numeroDeParadas)
		fatorProgresso := 0.8

		// Tenta encontrar um posto de recarga no caminho
		for {
			menorDist := math.MaxFloat64
			var postoMaisProximo consts.Posto
			var cidadeParada string
			encontrou := false

			// Verifica os postos de recarga das cidades
			for cidadeAtual, postos := range postosPorCidade {
				for _, posto := range postos {
					// Calcula a distância do posto para a posição atual
					dist := calcularDistancia(consts.Coordenadas{X: posto.X, Y: posto.Y}, posicaoAtual)
					if dist <= autonomiaAtual && dist >= fatorProgresso*distanciaPorParada && dist < menorDist {
						menorDist = dist
						postoMaisProximo = posto
						cidadeParada = cidadeAtual
						encontrou = true
					}
				}
			}

			// Se não encontrar um posto adequado, interrompe a busca
			if !encontrou {
				log.Printf("Nenhum posto encontrado para continuar a rota a partir de %s.", cidade)
				break
			}

			// Adiciona a parada ao mapa de paradas
			log.Printf("Parada adicionada: Cidade %s, Posto %+v", cidadeParada, postoMaisProximo)
			parada := consts.Parada{
				Cidade:       cidadeParada,
				PostoRecarga: postoMaisProximo,
			}
			paradasPorCidade[cidadeParada] = append(paradasPorCidade[cidadeParada], parada)
			posicaoAtual = consts.Coordenadas{X: postoMaisProximo.X, Y: postoMaisProximo.Y}
			autonomiaAtual = autonomiaTotal

			// Verifica se o carro pode chegar ao destino após a parada
			if calcularDistancia(destino, posicaoAtual) <= autonomiaAtual {
				log.Printf("Destino %s alcançado após a parada.", cidade)
				posicaoAtual = destino
				break
			}
		}
	}

	// Retorna as paradas necessárias para cada cidade
	return paradasPorCidade
}*/
