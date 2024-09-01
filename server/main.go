package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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

var db *sql.DB

type Capture struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PurchaseUnits []struct {
		Payments struct {
			Captures []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Amount struct {
					CurrencyCode string `json:"currency_code"`
					Value        string `json:"value"`
				} `json:"amount"`
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

func init() {
	godotenv.Load()

	if os.Getenv("BASE_URL") == "" ||
		os.Getenv("CLIENT_ID") == "" ||
		os.Getenv("CLIENT_SECRET") == "" ||
		os.Getenv("RETURN_URL") == "" ||
		os.Getenv("CANCEL_URL") == "" ||
		os.Getenv("PORT") == "" {
		log.Fatalf("Error 000: Missing credentials")
	}

	var err error
	db, err = sql.Open("postgres", "host="+os.Getenv("DB_HOST")+
		" port="+os.Getenv("DB_PORT")+
		" user="+os.Getenv("DB_USER")+
		" password="+os.Getenv("DB_PASS")+
		" dbname="+os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatalf("Error 001: Can't connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Error 001: Can't connect to database: %v", err)
	}

	log.Println("Established database connection")
}

func main() {
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

	defer func() {
		if db != nil {
			if err := db.Close(); err != nil {
				log.Fatalf("Error: Can't close database connection: %v", err)
			}
		}
	}()
	log.Println("Server shutdown gracefully.")
}

func Token() (string, error) {
	req, err := http.NewRequest("POST",
		os.Getenv("BASE_URL")+"/v1/oauth2/token",
		strings.NewReader(`grant_type=client_credentials`))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))

	raw, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v", err)
	}
	defer raw.Body.Close()

	var response struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(raw.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("Error decoding response: %v", err)
	}

	return response.AccessToken, nil
}

func RegisterOrder(capture Capture, directory string, editorData json.RawMessage) {
	var (
		// Payment
		id       string
		amount   string
		currency string
		pstatus  string
		date     time.Time
		// Website
		wstatus string
		due     time.Time
		name    string
		surname string
		email   string
		phone   string
		country string
	)

	id = capture.PurchaseUnits[0].Payments.Captures[0].ID
	amount = capture.PurchaseUnits[0].Payments.Captures[0].Amount.Value
	currency = capture.PurchaseUnits[0].Payments.Captures[0].Amount.CurrencyCode
	pstatus = capture.PurchaseUnits[0].Payments.Captures[0].Status
	date = capture.PurchaseUnits[0].Payments.Captures[0].CreateTime
	wstatus = "down"
	due = date.AddDate(1, 0, 0)
	name = capture.Payer.Name.GivenName
	surname = capture.Payer.Name.Surname
	email = capture.Payer.EmailAddress
	phone = capture.Payer.Phone.PhoneNumber.NationalNumber
	country = capture.Payer.Address.CountryCode

	var pkey int

	newSite := db.QueryRow(`SELECT id FROM sites WHERE folder = $1`, directory).Scan(&pkey)

	if newSite == sql.ErrNoRows {
		if err := db.QueryRow(
			`INSERT INTO sites (folder, status, due, name, sur, email, phone, code, raw)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`,
			directory, wstatus, due,
			name, surname, email, phone, country,
			editorData).Scan(&pkey); err != nil {
			log.Printf("Error: Could not register site to database: %v", err)
			return
		}
	} else {
		if err := db.QueryRow(
			`UPDATE sites SET due = due + INTERVAL '1 year'
			WHERE id = $1
			RETURNING id`,
			pkey).Scan(&pkey); err != nil {
			log.Fatalf("Error: Could not update due date: %v", err)
			return
		}
	}

	if _, err := db.Exec(
		`INSERT INTO payments (capture, site, amount, currency, date, status)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		id, pkey, amount, currency, date, pstatus); err != nil {
		log.Printf("Error: Could not register payment to database: %v", err)
		return
	}

	return
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	token, err := Token()
	if err != nil {
		http.Error(w,
			"Failed to get access token",
			http.StatusInternalServerError)
		return
	}

	type Amount struct {
		CurrencyCode string `json:"currency_code"`
		Value        string `json:"value"`
	}
	type PurchaseUnits struct {
		Amount Amount `json:"amount"`
	}
	type Address struct {
		CountryCode string `json:"country_code"`
	}
	type Paypal struct {
		Address Address `json:"address"`
	}
	type PaymentSource struct {
		Paypal Paypal `json:"paypal"`
	}
	type ApplicationContext struct {
		ShippingPreference string `json:"shipping_preference"`
	}
	type Order struct {
		Intent             string             `json:"intent"`
		PurchaseUnits      []PurchaseUnits    `json:"purchase_units"`
		PaymentSource      PaymentSource      `json:"payment_source"`
		ApplicationContext ApplicationContext `json:"application_context"`
	}

	// This payload will fill out defaults in PayPal
	// checkout window (CR code, no shipping, etc)
	order := Order{
		Intent: "CAPTURE",
		PurchaseUnits: []PurchaseUnits{{Amount: Amount{
			CurrencyCode: "USD",
			Value:        os.Getenv("PRICE"),
		}}},
		PaymentSource: PaymentSource{Paypal: Paypal{Address: Address{
			CountryCode: "CR",
		}}},
		ApplicationContext: ApplicationContext{
			ShippingPreference: "NO_SHIPPING",
		},
	}

	payload, err := json.Marshal(order)
	req, err := http.NewRequest("POST",
		os.Getenv("BASE_URL")+"/v2/checkout/orders",
		bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	raw, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w,
			"Failed to send request",
			http.StatusInternalServerError)
		return
	}
	defer raw.Body.Close()

	var response struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(raw.Body).Decode(&response); err != nil {
		http.Error(w,
			"Failed to decode response",
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	return
}

func CaptureOrder(w http.ResponseWriter, r *http.Request) {
	info, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w,
			"Failed to read request body",
			http.StatusInternalServerError)
		return
	}

	var cart struct {
		Directory  string          `json:"directory"`
		EditorData json.RawMessage `json:"editor_data"`
	}

	err = json.Unmarshal(info, &cart)
	if err != nil {
		http.Error(w,
			"Failed to parse request body",
			http.StatusBadRequest)
		return
	}

	directory := cart.Directory
	editorData := cart.EditorData
	if err != nil {
		http.Error(w,
			"Failed to parse request body",
			http.StatusBadRequest)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")
	orderID := parts[0]
	if orderID == "" {
		http.Error(w,
			"Failed to get orderID from client URL",
			http.StatusInternalServerError)
		return
	}

	token, err := Token()
	if err != nil {
		http.Error(w,
			"Failed to get access token",
			http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST",
		os.Getenv("BASE_URL")+"/v2/checkout/orders/"+orderID+"/capture",
		nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	raw, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w,
			"Failed to send request",
			http.StatusInternalServerError)
		return
	}
	defer raw.Body.Close()

	body, err := io.ReadAll(raw.Body)
	if err != nil {
		http.Error(w,
			"Failed to read response body",
			http.StatusInternalServerError)
		return
	}

	var capture Capture
	if err := json.Unmarshal(body, &capture); err != nil {
		http.Error(w,
			"Failed to decode response into capture",
			http.StatusInternalServerError)
		return
	}

	var receipt = struct {
		PurchaseUnits []struct {
			Payments struct {
				Captures []struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"captures"`
			} `json:"payments"`
		} `json:"purchase_units"`
	}{}

	if err := json.Unmarshal(body, &receipt); err != nil {
		http.Error(w,
			"Failed to decode response into receipt",
			http.StatusInternalServerError)
		return
	}

	RegisterOrder(capture, directory, editorData)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(receipt); err != nil {
		http.Error(w,
			"Failed to encode response",
			http.StatusInternalServerError)
		return
	}

	return
}
