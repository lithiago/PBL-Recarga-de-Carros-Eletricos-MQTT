package Rotas

import (
	consts "MQTT/utils/Constantes"
	"log"
	"strings"
)

// Calcula quantos quilômetros o carro pode andar com a bateria fornecida
func calcularAutonomia(consumoKW, capacidadeBateria float64) float64 {
	if consumoKW == 0 {
		return 0 // Evita divisão por zero
	}
	return capacidadeBateria / consumoKW
}

func GerarRotas(carro consts.Carro, rota []string, cidades map[string]consts.Coordenadas, todosOsPostos map[string][]consts.Posto) []consts.Parada {
	log.Println("🔄 Iniciando cálculo da rota com paradas automáticas...")

	posicaoAtual := consts.Coordenadas{X: carro.X, Y: carro.Y}
	bateriaAtual := carro.Bateria
	paradas := []consts.Parada{}

	for _, nomeCidade := range rota {
		destino := cidades[nomeCidade]

		// Se já estamos no destino, ignorar
		if posicaoAtual.X == destino.X && posicaoAtual.Y == destino.Y {
			continue
		}

		for {
			distancia := consts.CalcularDistancia(posicaoAtual, destino)
			autonomia := bateriaAtual / carro.Consumobateria
			log.Printf("📍 Tentando ir de (%.2f, %.2f) até %s. Distância: %.2f, Autonomia: %.2f", posicaoAtual.X, posicaoAtual.Y, nomeCidade, distancia, autonomia)

			// Se o destino é alcançável, simula a viagem e sai do loop
			if distancia <= autonomia {
				bateriaAtual -= distancia * carro.Consumobateria
				posicaoAtual = destino
				log.Printf("✅ Chegou diretamente em %s. Bateria restante: %.2f", nomeCidade, bateriaAtual)
				break
			}

			// Caso não seja alcançável, procurar melhor posto dentro da autonomia
			var melhorPosto *consts.Posto
			var menorDistancia float64 = 1e9

			for _, listaPostos := range todosOsPostos {
				for _, posto := range listaPostos {
					distanciaAtePosto := consts.CalcularDistancia(posicaoAtual, consts.Coordenadas{X: posto.X, Y: posto.Y})
					if distanciaAtePosto <= autonomia && distanciaAtePosto < menorDistancia {
						menorDistancia = distanciaAtePosto
						tmp := posto
						melhorPosto = &tmp
					}
				}
			}

			if melhorPosto == nil {
				log.Fatalf("❌ ERRO: Não há posto viável para recarga entre (%.2f, %.2f) e %s", posicaoAtual.X, posicaoAtual.Y, nomeCidade)
			}

			log.Printf("🔋 Parada necessária no posto: %s (%.2f, %.2f)", melhorPosto.Nome, melhorPosto.X, melhorPosto.Y)

			// Simula deslocamento até o posto
			distanciaAtePosto := consts.CalcularDistancia(posicaoAtual, consts.Coordenadas{X: melhorPosto.X, Y: melhorPosto.Y})
			bateriaAtual -= distanciaAtePosto * carro.Consumobateria
			bateriaAtual = carro.CapacidadeBateria // recarrega totalmente
			posicaoAtual = consts.Coordenadas{X: melhorPosto.X, Y: melhorPosto.Y}

			// Adiciona parada à lista
			paradas = append(paradas, consts.Parada{
				NomePosto: melhorPosto.Nome,
				IDPosto:   melhorPosto.Id,
				X:         melhorPosto.X,
				Y:         melhorPosto.Y,
			})
		}
	}

	log.Printf("🚗 Paradas planejadas (%d):", len(paradas))
	for i, p := range paradas {
		log.Printf("  [%d] %s (%s) - X: %.2f, Y: %.2f", i+1, p.NomePosto, p.IDPosto, p.X, p.Y)
	}

	return paradas
}

func GetRotasValidas(rotasPossiveis map[string][]string, trajeto consts.Trajeto) map[string][]string {
	inicio := strings.ToLower(trajeto.Inicio)
	destino := strings.ToLower(trajeto.Destino)

	mapaCompleto := make(map[string][]string)

	log.Printf("Início: %s, Destino: %s", inicio, destino)

	for nomeRota, rota := range rotasPossiveis {
		var indiceDestino int
		encontrouDestino := false

		// Encontrar índice do destino
		for j, cidade := range rota {
			if strings.ToLower(cidade) == destino {
				indiceDestino = j
				encontrouDestino = true
				break
			}
		}

		if !encontrouDestino {
			continue
		}

		// Verifica se o início está ANTES do destino
		for k, cidade := range rota[:indiceDestino+1] {
			if strings.ToLower(cidade) == inicio {
				if k == 0 { // Começa exatamente do início?
					log.Printf("✔ Rota válida encontrada: %s", nomeRota)
					mapaCompleto[nomeRota] = rota[:indiceDestino+1]
				}
				break
			}
		}
	}

	return mapaCompleto
}
