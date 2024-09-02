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

func main() {
	initialize()

	http.HandleFunc("/api/orders", CreateOrderHandler)
	http.HandleFunc("/api/orders/", CaptureOrderHandler)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	port := os.Getenv("PORT")

	go func() {
		log.Println("Starting server on " + port + "...")
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Error: 002: Can't start server: %v\n", err)
		}
	}()

	<-stop

	shutdown()
	log.Println("Server shutdown gracefully.")
}

func fail(w http.ResponseWriter, err error, notice string) {
	log.Printf("%s: %v", notice, err)
	http.Error(w, notice, http.StatusInternalServerError)
}

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID, err := CreateOrder()
	if err != nil {
		fail(w, err, "Failed to obtain orderID")
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
		fail(w, err, "Failed to read request body")
		return
	}
	var cart struct {
		Directory  string          `json:"directory"`
		EditorData json.RawMessage `json:"editor_data"`
	}
	err = json.Unmarshal(info, &cart)
	if err != nil {
		fail(w, err, "Failed to parse request body")
		return
	}
	directory := cart.Directory
	editorData := cart.EditorData

	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")
	orderID := parts[0]
	if orderID == "" {
		fail(w, err, "Failed to get orderID from client URL")
		return
	}

	capture, receipt, err := CaptureOrder(orderID)
	if err != nil {
		fail(w, err, "Failed to capture order")
		return
	}

	if err := RegisterSitePayment(capture, directory, editorData); err != nil {
		fail(w, err, "Failed to register '"+directory+"'in database")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(receipt); err != nil {
		fail(w, err, "Failed to encode response")
		return
	}

	return
}
