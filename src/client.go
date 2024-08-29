package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type CandyResponse struct {
	Thanks string `json:"thanks,omitempty"`
	Change int    `json:"change,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	// Параметры командной строки
	candyType := flag.String("k", "", "Two-letter abbreviation for the candy type")
	candyCount := flag.Int("c", 0, "Number of candies to buy")
	money := flag.Int("m", 0, "Amount of money given to the machine")
	flag.Parse()

	if *candyType == "" || *candyCount <= 0 || *money <= 0 {
		fmt.Println("Usage: ./candy-client -k CANDY_TYPE -c CANDY_COUNT -m MONEY")
		os.Exit(1)
	}

	// Загрузка CA сертификата
	caCert, err := ioutil.ReadFile("certs/ca-cert.pem")
	if err != nil {
		fmt.Println("Error loading CA cert:", err)
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Настройка TLS конфигурации для клиента
	clientCert, err := tls.LoadX509KeyPair("certs/client-cert.pem", "certs/client-key.pem")
	if err != nil {
		fmt.Println("Error loading client cert and key:", err)
		return
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{clientCert},
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Создание запроса
	order := map[string]interface{}{
		"money":      *money,
		"candyType":  *candyType,
		"candyCount": *candyCount,
	}
	orderJSON, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Error marshaling order:", err)
		return
	}

	req, err := http.NewRequest("POST", "https://localhost:8443/buy_candy", bytes.NewBuffer(orderJSON))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Отправка запроса
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// Обработка ответа
	var candyResp CandyResponse
	if err := json.NewDecoder(resp.Body).Decode(&candyResp); err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	if candyResp.Error != "" {
		fmt.Println(candyResp.Error)
	} else {
		fmt.Println(candyResp.Thanks)
		fmt.Println("Change: ", candyResp.Change)
	}
}
