package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	errTokenReq                  = "Error: paypal.go: Send token request"
	errTokenResp                 = "Error: paypal.go: Decode token response"
	errToken                     = "Error: paypal.go: Get access token"
	errCreateOrderReq            = "Error: paypal.go: Send CreateOrder request"
	errCreateOrderResp           = "Error: paypal.go: Decode CreateOrder response"
	errCaptureOrderToken         = "Error: paypal.go: Get access token for CaptureOrder"
	errCaptureOrderReq           = "Error: paypal.go: Send CaptureOrder request"
	errCaptureOrderResp          = "Error: paypal.go: Decode CaptureOrder response"
	errCaptureOrderDecodeCapture = "Error: paypal.go: Decode CaptureOrder capture"
	errCaptureOrderDecodeReceipt = "Error: paypal.go: Decode CaptureOrder receipt"
)

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

type Receipt struct {
	PurchaseUnits []struct {
		Payments struct {
			Captures []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

func Token() (string, error) {
	req, err := http.NewRequest("POST",
		os.Getenv("BASE_URL")+"/v1/oauth2/token",
		bytes.NewBufferString(`grant_type=client_credentials`))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))

	raw, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: %v", errTokenReq, err)
	}
	defer raw.Body.Close()

	var response struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(raw.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("%s: %v", errTokenResp, err)
	}

	return response.AccessToken, nil
}

func CreateOrder(amount string) (string, error) {
	token, err := Token()
	if err != nil {
		return "", fmt.Errorf("%s: %v", errToken, err)
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
			Value:        amount,
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
		return "", fmt.Errorf("%s: %v", errCreateOrderReq, err)
	}
	defer raw.Body.Close()

	if raw.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(raw.Body)
		return "", fmt.Errorf("PayPal API error: %s", string(body))
	}

	var response struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(raw.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("%s: %v", errCreateOrderResp, err)
	}

	return response.ID, nil
}

func CaptureOrder(orderID string) (Capture, Receipt, error) {
	token, err := Token()
	if err != nil {
		return Capture{}, Receipt{}, fmt.Errorf("%s: %v", errCaptureOrderToken, err)
	}

	req, err := http.NewRequest("POST",
		os.Getenv("BASE_URL")+"/v2/checkout/orders/"+orderID+"/capture",
		nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	raw, err := http.DefaultClient.Do(req)
	if err != nil {
		return Capture{}, Receipt{}, fmt.Errorf("%s: %v", errCaptureOrderReq, err)
	}
	defer raw.Body.Close()

	var capture Capture
	var receipt Receipt

	body, err := io.ReadAll(raw.Body)
	if err != nil {
		return Capture{}, Receipt{}, fmt.Errorf("%s: %v", errCaptureOrderResp, err)
	}

	if err := json.Unmarshal(body, &capture); err != nil {
		return Capture{}, Receipt{}, fmt.Errorf("%s: %v", errCaptureOrderDecodeCapture, err)
	}

	if err := json.Unmarshal(body, &receipt); err != nil {
		return Capture{}, Receipt{}, fmt.Errorf("%s: %v", errCaptureOrderDecodeReceipt, err)
	}

	return capture, receipt, nil
}
