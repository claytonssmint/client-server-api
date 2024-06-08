package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiURL          = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	dbName          = "cotacoes.db"
	apiTimeout      = 200 * time.Millisecond
	dbTimeout       = 10 * time.Millisecond
	httpServerPort  = ":8080"
	httpServerRoute = "/cotacao"
)

type Cotacao struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func getCotacao() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("falha ao buscar cotacao: %s", res.Status)
	}

	var cotacao Cotacao
	if err := json.NewDecoder(res.Body).Decode(&cotacao); err != nil {
		return "", err
	}

	return cotacao.USDBRL.Bid, nil
}

func saveCotacao(bid string) error {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := "INSERT INTO cotacoes (bid, timestamp) VALUES (?, ?)"
	_, err = db.ExecContext(ctx, query, bid, time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}
	return nil
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	bid, err := getCotacao()
	if err != nil {
		log.Printf("erro ao obter cotacao: %v", err)
		http.Error(w, "nao consegui buscar cotacao", http.StatusInternalServerError)
		return
	}

	if err := saveCotacao(bid); err != nil {
		log.Printf("erro ao save cotacao: %v", err)
		http.Error(w, "nao consegui salvar cotacao", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"bid": bid}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func main() {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatalf("erro ao abrir conexao com o banco de dados: %v", err)
	}
	defer db.Close()

	query := `
CREATE TABLE IF NOT EXISTS cotacoes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bid TEXT,
    timestamp TEXT
);
`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("erro ao criar tabela: %v", err)
	}

	http.HandleFunc(httpServerRoute, cotacaoHandler)
	log.Printf("iniciando servidor na porta %s%s", httpServerPort, httpServerRoute)
	log.Fatal(http.ListenAndServe(httpServerPort, nil))
}
