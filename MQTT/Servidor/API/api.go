package api

import (
	consts "MQTT/utils/Constantes"
	storage "MQTT/utils/storage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var postosMutex sync.Mutex

func ServerAPICommunication(arquivoPontos string) {

	r := gin.Default()

	r.GET("/postos", func(c *gin.Context) {
		postos, err := storage.GetPostosFromJSON(arquivoPontos)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(postos) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Nenhum posto encontrado"})
			return
		}
		// Retorna os postos como JSON
		c.JSON(http.StatusOK, postos)
	})

	r.GET("/postos/disponiveis", func(c *gin.Context) {
		postos, err := storage.GetPostosDisponiveis(arquivoPontos)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(postos) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "nenhum posto disponível encontrado"})
			return
		}
		c.JSON(http.StatusOK, postos)
	})

	r.PATCH("/postos/:id/adicionar", func(c *gin.Context) {
		id := c.Param("id")
		var carro consts.Carro
		if err := c.ShouldBindJSON(&carro); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
			return
		}

		postos, err := storage.GetPostosFromJSON(arquivoPontos)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var postoAtualizado *consts.Posto
		for _, p := range postos {
			if p.Id == id {
				// Verifica se já existe um carro na fila
				if len(p.Fila) > 0 {
					c.JSON(http.StatusConflict, gin.H{"error": "Já existe um carro na fila"})
					return
				}
				// Adiciona o carro à fila
				p.Fila = append(p.Fila, carro)
				postoAtualizado = p
				break
			}
		}

		if postoAtualizado == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Posto não encontrado"})
			return
		}

		if err := storage.AtualizarArquivo(arquivoPontos, postos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o arquivo JSON"})
			return
		}

		c.JSON(http.StatusOK, postoAtualizado)
	})

	r.PATCH("/postos/:id/remover", func(c *gin.Context) {
		id := c.Param("id")
		var carro consts.Carro
		if err := c.ShouldBindJSON(&carro); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
			return
		}

		postos, err := storage.GetPostosFromJSON(arquivoPontos)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var postoAtualizado *consts.Posto
		for _, p := range postos {
			if p.Id == id {
				for i, c := range p.Fila {
					if c.ID == carro.ID {
						p.Fila = append(p.Fila[:i], p.Fila[i+1:]...)
						postoAtualizado = p
						break
					}
				}
				break
			}
		}

		if postoAtualizado == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Posto ou carro não encontrado"})
			return
		}

		if err := storage.AtualizarArquivo(arquivoPontos, postos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o arquivo JSON"})
			return
		}

		c.JSON(http.StatusOK, postoAtualizado)
	})
	r.POST("/2pc/prepare", func(c *gin.Context) {
		var req struct {
			PostoID string       `json:"posto_id"`
			Carro   consts.Carro `json:"carro"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "abort", "error": "Dados inválidos"})
			return
		}

		postosMutex.Lock()
		defer postosMutex.Unlock()
		postos, err := storage.GetPostosFromJSON(arquivoPontos)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"result": "abort", "error": "Erro ao ler postos"})
			return
		}

		for _, p := range postos {
			if p.Id == req.PostoID {
				if len(p.Fila) > 0 || p.Pendente != nil {
					c.JSON(http.StatusOK, gin.H{"result": "abort"})
					return
				}
				// Marca como pendente
				p.Pendente = &req.Carro
				storage.AtualizarArquivo(arquivoPontos, postos)
				c.JSON(http.StatusOK, gin.H{"result": "ok"})
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"result": "abort", "error": "Posto não encontrado"})
	})
	r.POST("/2pc/commit", func(c *gin.Context) {
		var req struct {
			PostoID string       `json:"posto_id"`
			Carro   consts.Carro `json:"carro"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "abort", "error": "Dados inválidos"})
			return
		}
		postosMutex.Lock()
		defer postosMutex.Unlock()

		postos, err := storage.GetPostosFromJSON(arquivoPontos)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"result": "abort", "error": "Erro ao ler postos"})
			return
		}

		var postoAtualizado *consts.Posto
		for _, p := range postos {
			if p.Id == req.PostoID {
				// Só faz commit se o pendente for o mesmo carro
				if p.Pendente != nil && p.Pendente.ID == req.Carro.ID {
					p.Fila = append(p.Fila, req.Carro)
					p.Pendente = nil // limpa pendente
					postoAtualizado = p
				}
				break
			}
		}

		if postoAtualizado == nil {
			c.JSON(http.StatusNotFound, gin.H{"result": "abort", "error": "Posto não encontrado"})
			return
		}

		if err := storage.AtualizarArquivo(arquivoPontos, postos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"result": "abort", "error": "Erro ao atualizar o arquivo JSON"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": "committed"})
	})

	r.POST("/2pc/abort", func(c *gin.Context) {
		var req struct {
			PostoID string       `json:"posto_id"`
			Carro   consts.Carro `json:"carro"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "abort", "error": "Dados inválidos"})
			return
		}
		postosMutex.Lock()
		defer postosMutex.Unlock()

		postos, err := storage.GetPostosFromJSON(arquivoPontos)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"result": "abort", "error": "Erro ao ler postos"})
			return
		}

		for _, p := range postos {
			if p.Id == req.PostoID && p.Pendente != nil && p.Pendente.ID == req.Carro.ID {
				p.Pendente = nil
				storage.AtualizarArquivo(arquivoPontos, postos)
				break
			}
		}
		c.JSON(http.StatusOK, gin.H{"result": "aborted"})
	})

	r.POST("/reserva", func(c *gin.Context) {
		var req struct {
			Carro         consts.Carro             `json:"carro"`
			Participantes []consts.Participante2PC `json:"participantes"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
			return
		}
		log.Println("[API] Iniciando 2PC para adicionar carro aos postos...")
		err := TwoPhaseCommit(req.Participantes, req.Carro)
		if err != nil {
			log.Printf("[API] 2PC falhou: %v", err)
			c.JSON(http.StatusConflict, gin.H{"result": "2PC falhou", "error": err.Error()})
		} else {
			log.Println("[API] 2PC concluído com sucesso!")
			c.JSON(http.StatusOK, gin.H{"result": "2PC concluído com sucesso!"})
		}
	})

	// Inicia o servidor HTTP na porta 8080
	porta := os.Getenv("PORTA")
	if porta == "" {
		log.Fatalf("[SERVIDOR] Erro ao iniciar servidor HTTP com Gin: variável de ambiente PORTA não definida")
	}
	r.Run(":" + porta)
}

func ObterPostosDeOutroServidor(url string) ([]*consts.Posto, error) {
	//log.Printf("[SERVIDOR] Enviando requisição para %s/postos", url)

	// Cria a requisição HTTP GET
	resp, err := http.Get(url + "/postos/disponiveis")
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição para %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Verifica o status da resposta
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requisição falhou com status %d", resp.StatusCode)
	}

	// Decodifica o JSON da resposta
	var postos []consts.Posto
	if err := json.NewDecoder(resp.Body).Decode(&postos); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	// Convert to []*consts.Posto
	var postosPointers []*consts.Posto
	for i := range postos {
		postosPointers = append(postosPointers, &postos[i])
	}
	//log.Printf("[SERVIDOR] Postos recebidos de %s: %+v", url, postos)
	return postosPointers, nil

}

func TwoPhaseCommit(participantes []consts.Participante2PC, carro consts.Carro) error {
	payloadTemplate := `{"posto_id":"%s","carro":%s}`

	// Fase 1: Prepare
	okCount := 0
	for _, p := range participantes {
		carroJSON, _ := json.Marshal(carro)
		payload := fmt.Sprintf(payloadTemplate, p.PostoID, string(carroJSON))
		resp, err := http.Post(p.URL+"/2pc/prepare", "application/json", strings.NewReader(payload))
		if err != nil {
			log.Printf("[2PC] Erro ao enviar prepare para %s: %v", p.URL, err)
			break
		}
		defer resp.Body.Close()
		var res map[string]string
		json.NewDecoder(resp.Body).Decode(&res)
		if res["result"] == "ok" {
			okCount++
		} else {
			break
		}
	}

	// Se todos aceitaram, commit
	if okCount == len(participantes) {
		for _, p := range participantes {
			carroJSON, _ := json.Marshal(carro)
			payload := fmt.Sprintf(payloadTemplate, p.PostoID, string(carroJSON))
			http.Post(p.URL+"/2pc/commit", "application/json", strings.NewReader(payload))
		}
		log.Println("[2PC] Commit enviado para todos os participantes")
		return nil
	}

	// Se algum abortou, abort para todos
	for _, p := range participantes {
		carroJSON, _ := json.Marshal(carro)
		payload := fmt.Sprintf(payloadTemplate, p.PostoID, string(carroJSON))
		http.Post(p.URL+"/2pc/abort", "application/json", strings.NewReader(payload))
	}
	log.Println("[2PC] Abort enviado para todos os participantes")
	return fmt.Errorf("2PC abortado por algum participante")
}
