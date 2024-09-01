package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initialize()

	http.HandleFunc("/api/orders", CreateOrder)
	http.HandleFunc("/api/orders/", CaptureOrder)
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
