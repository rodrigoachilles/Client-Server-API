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
	Rate Rate `json:"USDBRL"`
}

type Rate struct {
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

type DollarRate struct {
	Dollar string `json:"dollar"`
}

const portServer = "8080"
const rateUrl = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

var (
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "INFO    | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "WARNING | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "ERROR   | ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serverHandler)
	InfoLogger.Println("Starting server on port " + portServer + "...")
	_ = http.ListenAndServe(":"+portServer, mux)
}

func serverHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/cotacao" {
		res.WriteHeader(http.StatusNotFound)
		WarningLogger.Println("Path [" + req.URL.Path + "] not found!")
		return
	}

	rate, err := findDollarRateOnTheInternet()
	if err != nil {
		ErrorLogger.Println(err)
		return
	}
	InfoLogger.Println("Dollar Rate found :", rate)

	rate, err = saveInDatabase(rate)
	if err != nil {
		ErrorLogger.Println(err)
	}
	InfoLogger.Println("Dollar Rate saved :", rate)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	dollarRate := DollarRate{Dollar: rate.Bid}
	_ = json.NewEncoder(res).Encode(dollarRate)
}

func findDollarRateOnTheInternet() (*Rate, error) {
	InfoLogger.Println("Finding Dollar Rate on the internet => url : " + rateUrl)

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
	return &result.Rate, nil
}

func saveInDatabase(rate *Rate) (*Rate, error) {
	InfoLogger.Println("Saving in file database sqlite...")

	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err = db.AutoMigrate(&Rate{})
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).Create(&rate).Error
	if err != nil {
		return nil, err
	}
	return rate, nil
}
