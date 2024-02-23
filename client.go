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

type DollarRateClient struct {
	Dollar string `json:"dollar"`
}

var (
	InfoLoggerClient  *log.Logger
	ErrorLoggerClient *log.Logger
)

func init() {
	InfoLoggerClient = log.New(os.Stdout, "INFO    | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	ErrorLoggerClient = log.New(os.Stdout, "ERROR   | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func main() {
	InfoLoggerClient.Println("Starting client...")

	dollar, err := findDollarRateInServer()
	if err != nil {
		ErrorLoggerClient.Println(err)
		return
	}

	saveInFile(dollar)
}

func findDollarRateInServer() (string, error) {
	InfoLoggerClient.Println("Finding Dollar Rate in the server => url : " + hostServer + resourceCotacao)

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
	var dollarRateClient DollarRateClient
	_ = json.Unmarshal(body, &dollarRateClient)
	_ = res.Body.Close()
	return dollarRateClient.Dollar, nil
}

func saveInFile(dollar string) {
	InfoLoggerClient.Println("Saving in the file 'cotacao.txt'...")
	InfoLoggerClient.Println("Dollar:", dollar)

	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	_, _ = f.Write([]byte("Dolar: " + dollar))
	_ = f.Close()
}
