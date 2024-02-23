package main

import (
	"context"
	"encoding/json"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Result struct {
	RateServer RateServer `json:"USDBRL"`
}

type RateServer struct {
	ID         int    `gorm:"primaryKey"`
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type DollarRateServer struct {
	Dollar string `json:"dollar"`
}

const portServer = "8080"
const rateUrl = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

var (
	InfoLoggerServer    *log.Logger
	WarningLoggerServer *log.Logger
	ErrorLoggerServer   *log.Logger
)

func init() {
	InfoLoggerServer = log.New(os.Stdout, "INFO    | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	WarningLoggerServer = log.New(os.Stdout, "WARNING | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	ErrorLoggerServer = log.New(os.Stdout, "ERROR   | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serverHandler)
	InfoLoggerServer.Println("Starting server on port " + portServer + "...")
	_ = http.ListenAndServe(":"+portServer, mux)
}

func serverHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/cotacao" {
		res.WriteHeader(http.StatusNotFound)
		WarningLoggerServer.Println("Path [" + req.URL.Path + "] not found!")
		return
	}

	rateServer, err := findDollarRateOnTheInternet()
	if err != nil {
		ErrorLoggerServer.Println(err)
		return
	}
	InfoLoggerServer.Println("Dollar Rate found :", rateServer)

	rateServer, err = saveInDatabase(rateServer)
	if err != nil {
		ErrorLoggerServer.Println(err)
	}
	InfoLoggerServer.Println("Dollar Rate saved :", rateServer)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	dollarRateServer := DollarRateServer{Dollar: rateServer.Bid}
	_ = json.NewEncoder(res).Encode(dollarRateServer)
}

func findDollarRateOnTheInternet() (*RateServer, error) {
	InfoLoggerServer.Println("Finding Dollar Rate on the internet => url : " + rateUrl)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", rateUrl, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var result Result
	_ = json.Unmarshal(body, &result)
	_ = res.Body.Close()
	return &result.RateServer, nil
}

func saveInDatabase(rateServer *RateServer) (*RateServer, error) {
	InfoLoggerServer.Println("Saving in file database sqlite...")

	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err = db.AutoMigrate(&RateServer{})
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).Create(&rateServer).Error
	if err != nil {
		return nil, err
	}
	return rateServer, nil
}
