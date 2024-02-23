package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const hostServer = "http://localhost:8080"
const resourceCotacao = "/cotacao"

type Rate struct {
	Dollar string `json:"dollar"`
}

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "INFO    | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "ERROR   | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func main() {
	InfoLogger.Println("Starting client...")

	dollar, err := findDollarRateInServer()
	if err != nil {
		ErrorLogger.Println(err)
		return
	}

	saveInFile(dollar)
}

func findDollarRateInServer() (string, error) {
	InfoLogger.Println("Finding Dollar Rate in the server => url : " + hostServer + resourceCotacao)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", hostServer+resourceCotacao, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var rate Rate
	_ = json.Unmarshal(body, &rate)
	_ = res.Body.Close()
	return rate.Dollar, nil
}

func saveInFile(dollar string) {
	InfoLogger.Println("Saving in the file 'cotacao.txt'...")
	InfoLogger.Println("Dollar:", dollar)

	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	_, _ = f.Write([]byte("Dolar: " + dollar))
	_ = f.Close()
}
