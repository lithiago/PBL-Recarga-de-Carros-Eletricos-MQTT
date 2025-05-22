package storage

import (
	consts "MQTT/utils/Constantes"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func GetPostosFromJSON(arquivoPontos string) ([]*consts.Posto, error) {
	filePath := arquivoPontos
	if filePath == "" {
		return nil, fmt.Errorf("ARQUIVO_JSON não definido")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo JSON: %v", err)
	}

	var mapa map[string][]consts.Posto
	if err := json.Unmarshal(data, &mapa); err != nil {
		return nil, fmt.Errorf("erro ao desserializar o JSON: %v", err)
	}

	cidade := os.Getenv("CIDADE")
	postos, ok := mapa[cidade]
	if !ok {
		return nil, fmt.Errorf("nenhum dado encontrado para a cidade: %s", cidade)
	}

	// Converte para []*consts.Posto
	var resultado []*consts.Posto
	for i := range postos {
		resultado = append(resultado, &postos[i])
	}

	return resultado, nil
}


func AtualizarArquivo(filePath string, postos []*consts.Posto) error {
	// Converte os postos para o formato de mapa esperado no JSON
	cidade := os.Getenv("CIDADE")
	if cidade == "" {
		return fmt.Errorf("CIDADE não definida")
	}

	mapa := map[string][]consts.Posto{
		cidade: make([]consts.Posto, len(postos)),
	}

	for i, posto := range postos {
		mapa[cidade][i] = *posto
	}

	// Serializa o mapa para JSON
	data, err := json.MarshalIndent(mapa, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar os dados para JSON: %v", err)
	}

	// Escreve os dados no arquivo
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("erro ao escrever no arquivo JSON: %v", err)
	}

	return nil
}


func GetPostosDisponiveis(arquivoPontos string) ([]*consts.Posto, error) {
	postos, err := GetPostosFromJSON(arquivoPontos)
	if err != nil {
		return nil, err
	}

	var postosDisponiveis []*consts.Posto
	for _, posto := range postos {
		if len(posto.Fila) == 0 { // Verifica se a fila está vazia
			postosDisponiveis = append(postosDisponiveis, posto)
		}
	}

	if len(postosDisponiveis) == 0 {
		return nil, fmt.Errorf("nenhum posto disponível encontrado")
	}

	return postosDisponiveis, nil
}


func CarregarPostos() []consts.Posto {
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
	return postos
}
