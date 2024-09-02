package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	msgServerStart    = "Msg: main.go: Starting server"
	msgServerShutdown = "Msg: main.go: Server shutdown gracefully"
	errServerStart    = "Fatal: main.go: Start server"
	errReadBody       = "Error: main.go: Read request body"
	errParseBody      = "Error: main.go: Parse request body"
	errGetOrderID     = "Error: main.go: Get orderID from client URL"
	errCaptureOrder   = "Error: main.go: Capture order"
	errRegisterSite   = "Error: main.go: Register site in database"
	errEncodeResponse = "Error: main.go: Encode response"
	errCreateOrder    = "Error: main.go: Obtain orderID"
)

func main() {
	initialize()

	http.HandleFunc("/api/orders", CreateOrderHandler)
	http.HandleFunc("/api/orders/", CaptureOrderHandler)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	port := os.Getenv("PORT")

	go func() {
		msg(msgServerStart + ": " + port + "...")
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fatal(err, errServerStart)
		}
	}()

	<-stop

	shutdown()
	msg(msgServerShutdown)
}

func msg(notice string) {
	log.Println(notice)
}

func fail(w http.ResponseWriter, err error, notice string) {
	log.Printf("%s: %v", notice, err)
	http.Error(w, notice, http.StatusInternalServerError)
}

func fatal(err error, notice string) {
	shutdown()
	log.Fatalf("%s: %v", notice, err)
}

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID, err := CreateOrder()
	if err != nil {
		fail(w, err, errCreateOrder)
		return
	}

	var response struct {
		ID string `json:"id"`
	}
	response.ID = orderID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	return
}

func CaptureOrderHandler(w http.ResponseWriter, r *http.Request) {
	var cart struct {
		Directory  string          `json:"directory"`
		EditorData json.RawMessage `json:"editor_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
		fail(w, err, errReadBody)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")
	orderID := parts[0]
	if orderID == "" {
		fail(w, nil, errGetOrderID)
		return
	}

	capture, receipt, err := CaptureOrder(orderID)
	if err != nil {
		fail(w, err, errCaptureOrder)
		return
	}

	if err := RegisterSitePayment(capture, cart.Directory, cart.EditorData); err != nil {
		fail(w, err, errRegisterSite+": "+cart.Directory)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(receipt); err != nil {
		fail(w, err, errEncodeResponse)
		return
	}

	return
}
