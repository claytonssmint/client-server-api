package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	serverURL      = "http://localhost:8080/cotacao"
	requestTimeout = 300 * time.Millisecond
	outputFile     = "cotacao.txt"
)

type Response struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		log.Fatalf("erro ao criar a requisicao: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("erro ao fazer a requisicao: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("a resposta veio diferente de status 200: %s", resp.Status)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("erro ao decodificar a resposta: %v", err)
	}

	content := fmt.Sprintf("Dólar: %s", response.Bid)
	if err := ioutil.WriteFile(outputFile, []byte(content), 0644); err != nil {
		log.Fatalf("error ao gravar no arquivo")
	}

	log.Printf("cotacao salva com sucesso %s", outputFile)
}
