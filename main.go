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
	// "time"

	"github.com/joho/godotenv"
)

type OrderData struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PurchaseUnits []struct {
		Payments struct {
			Captures []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

type SubscriptionData struct {
	ID               string    `json:"id"`
	Status           string    `json:"status"`
	// StatusUpdateTime time.Time `json:"status_update_time"`
	PlanID           string    `json:"plan_id"`
	PlanOverridden   bool      `json:"plan_overridden"`
	// StartTime        time.Time `json:"start_time"`
	Quantity         string    `json:"quantity"`
	ShippingAmount   struct {
		CurrencyCode string `json:"currency_code"`
		Value        string `json:"value"`
	} `json:"shipping_amount"`
	Subscriber struct {
		Name struct {
			GivenName string `json:"given_name"`
			Surname   string `json:"surname"`
		} `json:"name"`
		EmailAddress    string `json:"email_address"`
		PayerID         string `json:"payer_id"`
		ShippingAddress struct {
			Name struct {
				FullName string `json:"full_name"`
			} `json:"name"`
			Address struct {
				AddressLine1 string `json:"address_line_1"`
				AddressLine2 string `json:"address_line_2"`
				AdminArea2   string `json:"admin_area_2"`
				AdminArea1   string `json:"admin_area_1"`
				PostalCode   string `json:"postal_code"`
				CountryCode  string `json:"country_code"`
			} `json:"address"`
		} `json:"shipping_address"`
	} `json:"subscriber"`
	// CreateTime time.Time `json:"create_time"`
	Links      []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

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
	http.HandleFunc("/api/orders/", CaptureOrder)
	http.HandleFunc("/api/paypal/create-subscription", CreateSubscription)
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

func CaptureOrder(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")
	orderID := parts[0]

	client := &http.Client{}
	req, err := http.NewRequest("POST", baseURL+"/v2/checkout/orders/"+orderID+"/capture", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	token, err := Token()
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Create an instance of AutoGenerated
	var result OrderData

	// Decode the response into the AutoGenerated struct
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	// Now, `result` contains the entire structured response
	// You can send the whole `result` back to the client, or you can selectively send fields.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	log.Printf("asked to create sub")

	planID := os.Getenv("PLAN_ID")
	returnUrl := "https://suckless.org"
	cancelUrl := "https://suckless.org"

	log.Printf("This is the planid: %s", planID)

	token, err := Token()
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	log.Printf("This is the token: %s", token)

	body := map[string]interface{}{
		"plan_id": planID,
		"application_context": map[string]string{
			"shipping_preference": "NO_SHIPPING",
			"return_url":          returnUrl,
			"cancel_url":          cancelUrl,
		},
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Creating request")
	req, err := http.NewRequest("POST", baseURL+"/v1/billing/subscriptions", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Prefer", "return=representation")

	log.Printf("Sending request")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	log.Printf("Request sent")

	// Create an instance of AutoGenerated
	// var result SubscriptionData
	var result map[string]interface{}

	// Decode the response into the AutoGenerated struct
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}
	log.Printf("Raw JSON Response: %v", result)

	// Now, `result` contains the entire structured response
	// You can send the whole `result` back to the client, or you can selectively send fields.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
	log.Printf("sent response to client")

}
