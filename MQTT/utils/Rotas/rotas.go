package Rotas

import (
	consts "MQTT/utils/Constantes"
	"log"
)

// Calcula quantos quilômetros o carro pode andar com a bateria fornecida
func calcularAutonomia(consumoKW, capacidadeBateria float64) float64 {
	if consumoKW == 0 {
		return 0 // Evita divisão por zero
	}
	return capacidadeBateria / consumoKW
}

/* func GerarRotas(carro consts.Carro, rota []string, cidades map[string]consts.Coordenadas, postosPorCidade map[string][]consts.Posto) map[string][]consts.Parada {
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

	logParadas(paradasPorCidade)

	return paradasPorCidade
}


func logParadas(paradas map[string][]consts.Parada) {
	log.Println("=== PARADAS DEFINIDAS NA ROTA ===")
	for cidade, lista := range paradas {
		log.Printf("[CIDADE] %s", cidade)
		for i, parada := range lista {
			p := parada.PostoRecarga
			log.Printf("  -> [%d] Posto: %s (ID: %s)", i+1, p.Nome, p.Id)
			log.Printf("     Coordenadas: (X: %.2f, Y: %.2f)", p.X, p.Y)
			log.Printf("     Custo/kWh: R$ %.2f", p.CustoKW)
			log.Printf("     Carros na fila: %d", len(p.Fila))
		}
	}
	log.Println("==================================")
}


*/

/* func GerarRotasOtimizadas (carro consts.Carro, rotas[]string, cidades map[string]consts.Coordenadas, todosOsPostos []consts.Posto)map[string][]consts.Parada{
	log.Println("Criando rotas otmizadas...")
	autonomiaTotal := calcularAutonomia(carro.Consumobateria, carro.CapacidadeBateria)
	log.Printf("Autonomia total: %.2f km", autonomiaTotal)


	posicaoCarro := consts.Coordenadas{X: carro.X, Y: carro.Y}
	grafo := grafoLib.CriarGrafo(cidades, todosOsPostos, autonomiaTotal)
	paradasPorRota := make(map[string][]consts.Parada)
	for _, cidadeDestino := range rotas {
		coordenadasDestino := cidades[cidadeDestino]
		log.Printf("Coordenadas destino: %+v", coordenadasDestino)
		destID := strings.ToLower(strings.ReplaceAll(cidadeDestino, " ", ""))

		// Um nó temporário para a posição atual do carro
		idCarro := "carro_atual"
		grafo.Nos[idCarro] = grafoLib.No{
			ID:         idCarro,
			Coordenada: posicaoCarro,
		}
		for idDestino, no := range grafo.Nos{
			if idDestino == idCarro{
				continue
			}
			distancia:= CalcularDistancia(posicaoCarro, no.Coordenada)
			if distancia <= autonomiaTotal{
				grafo.Arestas[idCarro] = append(grafo.Arestas[idCarro], grafoLib.Aresta{
					Destino:   idDestino,
					Distancia: distancia,
				})
			}
		}


		// Executa o algoritmo de caminho mínimo (Dijkstra ou BFS com contagem de saltos)
	paradas, err := grafoLib.EncontrarCaminhoComMinimoDeParadasNoGrafo(grafo, idCarro, destID)
	if err != nil {
		log.Printf("Erro ao gerar rota até %s: %v", cidadeDestino, err)
		continue
	}

	paradasPorRota[cidadeDestino] = paradas
	}
} */

func GerarRotas(carro consts.Carro, rota []string, cidades map[string]consts.Coordenadas, todosOsPostos map[string][]consts.Posto) []consts.Parada {
	log.Println("Criando rotas...")
	log.Printf("Bateria: %+v", carro.Bateria)
	log.Printf("Rotas: %v", rota)

	autonomiaTotal := calcularAutonomia(carro.Consumobateria, carro.CapacidadeBateria)
	log.Printf("Autonomia total: %.2f km", autonomiaTotal)

	posicaoCarro := consts.Coordenadas{X: carro.X, Y: carro.Y}
	paradasPorCidade := []consts.Parada{}

	for _, cidade := range rota {
		destino := cidades[cidade]

			// Se já estiver na cidade, segue para o próximo destino
			if posicaoCarro.X == destino.X && posicaoCarro.Y == destino.Y {
				log.Printf("O carro já está em %s.", cidade)
				continue
			}
		

		for {
			distanciaAteDestino := consts.CalcularDistancia(posicaoCarro, destino)
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

			for cidadeAtual, postos := range todosOsPostos {
				for _, posto := range postos {
					distanciaAtePosto := consts.CalcularDistancia(posicaoCarro, consts.Coordenadas{X: posto.X, Y: posto.Y})
					distanciaPostoAteDestino := consts.CalcularDistancia(consts.Coordenadas{X: posto.X, Y: posto.Y}, destino)

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
				NomePosto:       melhorPosto.Nome,
				IDPosto: melhorPosto.Id,
				X:         melhorPosto.X,
				Y:         melhorPosto.Y,
			}
			paradasPorCidade = append(paradasPorCidade, parada)

			// Atualiza a posição do carro para o posto escolhido
			posicaoCarro = consts.Coordenadas{X: melhorPosto.X, Y: melhorPosto.Y}
		}
	}

	return paradasPorCidade
}

