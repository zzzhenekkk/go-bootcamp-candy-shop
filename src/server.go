package main

/*
#include <stdlib.h>
#include "cow.c"  // Включение C файла напрямую

// Прототип функции для использования в Go
char* ask_cow(char phrase[]);
*/
import "C"
import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"unsafe"
)

var CandyPrices = map[string]int{
	"CE": 10,
	"AA": 15,
	"NT": 17,
	"DE": 21,
	"YR": 23,
}

type CandyResponse struct {
	Thanks string `json:"thanks,omitempty"`
	Change int    `json:"change,omitempty"`
	Error  string `json:"error,omitempty"`
}

func askCow(phrase string) string {
	cPhrase := C.CString(phrase)
	defer C.free(unsafe.Pointer(cPhrase))

	cResult := C.ask_cow(cPhrase)
	defer C.free(unsafe.Pointer(cResult))

	return C.GoString(cResult)
}

func buyCandyHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var order map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&order); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	money, moneyOk := order["money"].(float64)
	candyType, candyTypeOk := order["candyType"].(string)
	candyCount, candyCountOk := order["candyCount"].(float64)

	if !moneyOk || !candyTypeOk || !candyCountOk || candyCount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CandyResponse{Error: "Invalid type or count of candy"})
		return
	}

	price, exists := CandyPrices[candyType]
	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CandyResponse{Error: "Invalid candy type"})
		return
	}

	totalCost := price * int(candyCount)

	if int(money) < totalCost {
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(CandyResponse{Error: fmt.Sprintf("You need %d more money!", totalCost-int(money))})
	} else {
		change := int(money) - totalCost
		thanksMessage := askCow("Thank you!")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CandyResponse{Thanks: thanksMessage, Change: change})
	}
}

func main() {
	// Загрузка CA сертификата
	caCert, err := ioutil.ReadFile("certs/ca-cert.pem")
	if err != nil {
		fmt.Println("Error loading CA cert:", err)
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Настройка TLS конфигурации с mTLS
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
	}

	http.HandleFunc("/buy_candy", buyCandyHandler)

	fmt.Println("HTTPS Server with mTLS is running on port 8443...")
	err = server.ListenAndServeTLS("certs/server-cert.pem", "certs/server-key.pem")
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
