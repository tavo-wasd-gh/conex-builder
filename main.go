package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	baseURL      string
	clientID     string
	clientSecret string
	planID       string
	returnUrl    string
	cancelUrl    string

	exists bool
	err error

	dbHost string
	dbPort string
	dbUser string
	dbPass string
	dbName string
	db     *sql.DB
	query  string
)

func init() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Set variables
	baseURL = os.Getenv("BASE_URL")
	clientID = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	planID = os.Getenv("PLAN_ID")
	returnUrl = os.Getenv("RETURN_URL")
	cancelUrl = os.Getenv("CANCEL_URL")
	// DB creds
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbUser = os.Getenv("DB_USER")
	dbPass = os.Getenv("DB_PASS")
	dbName = os.Getenv("DB_NAME")

	// Error if empty
	if baseURL == "" || clientID == "" || clientSecret == "" || planID == "" || returnUrl == "" || cancelUrl == "" ||
		dbHost == "" || dbPort == "" || dbUser == "" || dbPass == "" || dbName == "" {
		log.Fatalf("Error setting credentials")
	}

	// Connect to DB
	var err error
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Ping DB
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
}

type CreateOrderResponse struct {
	ID string `json:"id"`
}

type OrderResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PurchaseUnits []struct {
		Payments struct {
			Captures []struct {
				ID         string    `json:"id"`
				Status     string    `json:"status"`
				CreateTime time.Time `json:"create_time"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
	Payer struct {
		Name struct {
			GivenName string `json:"given_name"`
			Surname   string `json:"surname"`
		} `json:"name"`
		EmailAddress string `json:"email_address"`
		Phone        struct {
			PhoneType   string `json:"phone_type"`
			PhoneNumber struct {
				NationalNumber string `json:"national_number"`
			} `json:"phone_number"`
		} `json:"phone"`
		Address struct {
			CountryCode string `json:"country_code"`
		} `json:"address"`
	} `json:"payer"`
}

type SubscriptionResponse struct {
	Status           string    `json:"status"`
	StatusUpdateTime time.Time `json:"status_update_time"`
	StartTime        time.Time `json:"start_time"`
	Subscriber struct {
		Name struct {
			GivenName string `json:"given_name"`
			Surname   string `json:"surname"`
		} `json:"name"`
		EmailAddress    string `json:"email_address"`
	} `json:"subscriber"`
	CreateTime       time.Time `json:"create_time"`
}

type Cart struct {
	Directory string `json:"directory"`
}

func main() {
	http.HandleFunc("/api/order", CreateOrder)
	http.HandleFunc("/api/order/", CaptureOrder)
	http.HandleFunc("/api/paypal/subscribe", CreateSubscription)
	http.HandleFunc("/api/paypal/subscribe/", CaptureSubscription)
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
	// Create
	req, err := http.NewRequest("POST", baseURL+"/v1/oauth2/token", strings.NewReader(`grant_type=client_credentials`))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %v", err)
	}

	// Send
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Decode
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("Error decoding response: %v", err)
	}

	// Return
	return result.AccessToken, nil
}

func RegisterOrder(order OrderResponse, directory string) {
	var (
		capture string
		status  string
		name    string
		surname string
		email   string
		phone   string
		country string
		date    time.Time
	)

	for _, Unit := range order.PurchaseUnits {
		for _, Capture := range Unit.Payments.Captures {
			capture = Capture.ID
			status = Capture.Status
			date = Capture.CreateTime
		}
	}

	name = order.Payer.Name.GivenName
	surname = order.Payer.Name.Surname
	email = order.Payer.EmailAddress
	phone = order.Payer.Phone.PhoneNumber.NationalNumber
	country = order.Payer.Address.CountryCode

	// Register Payment
	_, err = db.Exec(`INSERT INTO payments (id,      client, directory, status, step,         date) VALUES ($1, $2, $3, $4, $5, $6);`,
		capture, email, directory, status, "REGISTERED", date)
	if err != nil {
		fmt.Printf("$v", err) // TODO consider logging in server
	}

	// Register Client
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM clients WHERE email = $1);`, email).Scan(&exists)
	if err != nil {
		fmt.Printf("$v", err) // TODO consider logging in server
	}
	if !exists {
		_, err = db.Exec(`INSERT INTO clients (email, name, surname, phone, country) VALUES ($1, $2, $3, $4, $5);`,
			email, name, surname, phone, country)
		if err != nil {
			fmt.Printf("$v", err) // TODO consider logging in server
		}
	}

	// Register Site
	_, err = db.Exec(`INSERT INTO sites (directory, client, status,  ends) VALUES ($1, $2, $3, $4);`,
		directory, email, "ACTIVE", date.AddDate(1, 0, 0)) // Ends a year later
	if err != nil {
		fmt.Printf("$v", err) // TODO consider logging in server
	}
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
		"payment_source": {
			"paypal": {
				"address" : {
					"country_code": "CR"
				}
			}
		},
		"application_context": {
			"shipping_preference": "NO_SHIPPING"
		}
	}`

	// Create
	req, err := http.NewRequest("POST", baseURL+"/v2/checkout/orders", bytes.NewBuffer([]byte(data)))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Send
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode
	var result CreateOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": result.ID})
	return
}

func CaptureOrder(w http.ResponseWriter, r *http.Request) {
	// Read body from order
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// Parse to get directory
	var cart Cart
	err = json.Unmarshal(body, &cart)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	directory := cart.Directory

	// Get orderID
	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")
	orderID := parts[0]
	if orderID == "" {
		http.Error(w, "Failed to get orderID from client URL", http.StatusInternalServerError)
		return
	}

	token, err := Token()
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	// Create
	req, err := http.NewRequest("POST", baseURL+"/v2/checkout/orders/"+orderID+"/capture", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Check if directory already exists
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sites WHERE directory = $1 LIMIT 1);", directory).Scan(&exists)
	if err != nil {
		http.Error(w, "Failed to check directory ID against database", http.StatusBadRequest)
		return
	}
	if exists {
		http.Error(w, "This directory ID is already taken", http.StatusBadRequest)
		return
	}

	// Send, PAYMENT MADE HERE
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode
	var order OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	RegisterOrder(order, directory)

	// Respond
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	// Read body from order
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// Parse to get directory
	var cart Cart
	err = json.Unmarshal(body, &cart)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	directory := cart.Directory

	token, err := Token()
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	payload := map[string]interface{}{
		"plan_id": planID,
		"application_context": map[string]string{
			"shipping_preference": "NO_SHIPPING",
			"return_url":          returnUrl,
			"cancel_url":          cancelUrl,
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Create request
	log.Printf("Creating request")
	req, err := http.NewRequest("POST", baseURL+"/v1/billing/subscriptions", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Check if directory already exists
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sites WHERE directory = $1 LIMIT 1);", directory).Scan(&exists)
	if err != nil {
		http.Error(w, "Failed to check directory ID against database", http.StatusBadRequest)
		return
	}
	if exists {
		http.Error(w, "This directory ID is already taken", http.StatusBadRequest)
		return
	}

	// Send
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Prefer", "return=representation")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	// Respond
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
	log.Printf("sent response to client")
}

// Capture just like CaptureOrder, but with response from paypal
// https://developer.paypal.com/docs/api/subscriptions/v1/#subscriptions_get
func CaptureSubscription(w http.ResponseWriter, r *http.Request) {
	// Read body from order
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// Parse to get directory
	var cart Cart
	err = json.Unmarshal(body, &cart)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	// directory := cart.Directory

	// Get subID
	path := strings.TrimPrefix(r.URL.Path, "/api/subscribe/")
	parts := strings.Split(path, "/")
	subID := parts[0]
	if subID == "" {
		http.Error(w, "Failed to get subID from client URL", http.StatusInternalServerError)
		return
	}
}
