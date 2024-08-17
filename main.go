package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"os"
	"fmt"
	"os/signal"
	"syscall"
	"bytes"

	"github.com/joho/godotenv"
)

var (
	baseURL = "https://api-m.sandbox.paypal.com"
)

func init() {
	if err := godotenv.Load() ; err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	// Handlers
	http.HandleFunc("/api/orders", CreateOrder)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	// Channel to listen for signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Run the server in a goroutine so that it doesn't block
	go func() {
		log.Println("Starting server on :8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Could not listen on :8080: %v\n", err)
		}
	}()

	<-stop // Shutdown signal recieved
	log.Println("Server shutdown gracefully.")
}

func Token() (string, error) {
	// Create request
	req, err := http.NewRequest("POST", baseURL+"/v1/oauth2/token", strings.NewReader(`grant_type=client_credentials`))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %v", err)
	}

	// Make POST req, should return JSON with AccessToken
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Decode response into result
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("Error decoding response: %v", err)
	}

	return result.AccessToken, nil
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	token, err := Token()
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	data := `{
		"intent": "CAPTURE",
		"purchase_units": [{
			"amount": {
				"currency_code": "USD",
				"value": "20.00"
			}
		}],
		"application_context": {
			"shipping_preference": "NO_SHIPPING"
		}
	}`

	req, err := http.NewRequest("POST", baseURL+"/v2/checkout/orders", bytes.NewBuffer([]byte(data)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	if id, ok := result["id"].(string); ok {
		json.NewEncoder(w).Encode(map[string]string{"id": id})
		return
	} else {
		http.Error(w, "Order ID not found", http.StatusInternalServerError)
		return
	}
}
