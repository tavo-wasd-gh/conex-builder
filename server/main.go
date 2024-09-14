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
	errAuthGen        = "Error: main.go: Gen and register auth"
	errAuthEmail      = "Error: main.go: Send auth email"
	errAuthValidate   = "Error: main.go: Validate changes"
	errUpdateSite     = "Error: main.go: Updating site data"
)

func main() {
	initialize()

	http.HandleFunc("/api/orders", CreateOrderHandler)
	http.HandleFunc("/api/orders/", CaptureOrderHandler)
	http.HandleFunc("/api/update", UpdateSiteHandler)
	http.HandleFunc("/api/confirm", ConfirmChangesHandler)
	http.HandleFunc("/api/directory/", VerifyDirectoryHandler)
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

func httpErrorAndLog(w http.ResponseWriter,
	err error, notice string, client string,
) {
	log.Printf("%s: %v", notice, err)
	http.Error(w, client, http.StatusInternalServerError)
}

func fatal(err error, notice string) {
	shutdown()
	log.Fatalf("%s: %v", notice, err)
}

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var cart struct {
		Directory string `json:"directory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
		httpErrorAndLog(w, err, errReadBody, "Error decoding response")
		return
	}
	if err := AvailableSite(cart.Directory) ; err != nil {
		http.Error(w, "Site already exists", http.StatusConflict)
		log.Printf("%s: %v", "Site already exists", err)
		return
	}

	orderID, err := CreateOrder()
	if err != nil {
		httpErrorAndLog(w, err, errCreateOrder, "Error creating order")
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
	errClientNotice := "Error capturing order"

	var cart struct {
		Directory  string          `json:"directory"`
		EditorData json.RawMessage `json:"editor_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
		httpErrorAndLog(w, err, errReadBody, errClientNotice)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")
	orderID := parts[0]
	if orderID == "" {
		httpErrorAndLog(w, nil, errGetOrderID, errClientNotice)
		return
	}

	capture, receipt, err := CaptureOrder(orderID)
	if err != nil {
		httpErrorAndLog(w, err, errCaptureOrder, errClientNotice)
		return
	}

	if err := RegisterSitePayment(capture, cart.Directory, cart.EditorData); err != nil {
		httpErrorAndLog(w, err, errRegisterSite+": "+cart.Directory, errClientNotice)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(receipt); err != nil {
		httpErrorAndLog(w, err, errEncodeResponse, errClientNotice)
		return
	}

	return
}

func UpdateSiteHandler(w http.ResponseWriter, r *http.Request) {
	errClientNotice := "Error handling update request"

	var cart struct {
		Directory string `json:"directory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
		httpErrorAndLog(w, err, errReadBody, errClientNotice)
		return
	}

	code := GenerateCode()

	email, err := UpdateSiteAuth(cart.Directory, code)
	if err != nil {
		httpErrorAndLog(w, err, errAuthGen, errClientNotice)
		return
	}

	if err := SendAuthEmail(email, code); err != nil {
		httpErrorAndLog(w, err, errAuthEmail, errClientNotice)
		return
	}

	return
}

func ConfirmChangesHandler(w http.ResponseWriter, r *http.Request) {
	errClientNotice := "Error handling confirm changes request"

	var cart struct {
		Directory  string          `json:"directory"`
		Code       string          `json:"auth_code"`
		EditorData json.RawMessage `json:"editor_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
		httpErrorAndLog(w, err, errReadBody, errClientNotice)
		return
	}

	pkey, err := ValidateSiteAuth(cart.Directory, cart.Code)
	if err != nil {
		httpErrorAndLog(w, err, errAuthValidate, errClientNotice)
		return
	}

	if err := UpdateSite(pkey, cart.EditorData); err != nil {
		httpErrorAndLog(w, err, errUpdateSite, errClientNotice)
		return
	}

	return
}

func VerifyDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	errClientNotice := "Error verifying directory against db"

	path := strings.TrimPrefix(r.URL.Path, "/api/directory/")
	parts := strings.Split(path, "/")
	folder := parts[0]
	if folder == "" {
		httpErrorAndLog(w, nil, "Error getting directory", errClientNotice)
		return
	}

	var response struct {
		Exists bool `json:"exists"`
	}

	err := AvailableSite(folder)
	if err != nil {
		response.Exists = true
	} else {
		response.Exists = false
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	return
}
