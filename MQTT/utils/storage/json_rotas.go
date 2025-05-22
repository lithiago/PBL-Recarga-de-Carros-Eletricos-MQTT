package storage

import (
	consts "MQTT/utils/Constantes"
	"encoding/json"
	"os"
)

func LerRotas() consts.DadosRotas{
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