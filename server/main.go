package main

import (
	"encoding/json"
	"io"
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
		log.Println(msgServerStart + ": " + port + "...")
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fatal(err, errServerStart)
		}
	}()

	<-stop

	shutdown()
	log.Println(msgServerShutdown)
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
	info, err := io.ReadAll(r.Body)
	if err != nil {
		fail(w, err, errReadBody)
		return
	}
	var cart struct {
		Directory  string          `json:"directory"`
		EditorData json.RawMessage `json:"editor_data"`
	}
	err = json.Unmarshal(info, &cart)
	if err != nil {
		fail(w, err, errParseBody)
		return
	}
	directory := cart.Directory
	editorData := cart.EditorData

	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")
	orderID := parts[0]
	if orderID == "" {
		fail(w, err, errGetOrderID)
		return
	}

	capture, receipt, err := CaptureOrder(orderID)
	if err != nil {
		fail(w, err, errCaptureOrder)
		return
	}

	if err := RegisterSitePayment(capture, directory, editorData); err != nil {
		fail(w, err, errRegisterSite+": "+directory)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(receipt); err != nil {
		fail(w, err, errEncodeResponse)
		return
	}

	return
}
