package main

import (
	"context"
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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	// Limits
	maxUploadFileSize = 52428800    // 50MB
	maxBucketSize     = 10737418240 // 10GB
	// Messages
	msgClosingDBConn      = "Msg: init.go: Closing database connection"
	msgDBConn             = "Msg: init.go: Established database connection"
	errDBConn             = "Fatal: init.go: Connect to database"
	errDBPing             = "Fatal: init.go: Ping database"
	errClosingDBConn      = "Fatal: init.go: Closing database connection"
	errMissingCredentials = "Fatal: init.go: Credentials"
	msgServerStart        = "Msg: main.go: Starting server"
	msgServerShutdown     = "Msg: main.go: Server shutdown gracefully"
	errServerStart        = "Fatal: main.go: Start server"
	errReadBody           = "Error: main.go: Read request body"
	errParseBody          = "Error: main.go: Parse request body"
	errGetOrderID         = "Error: main.go: Get orderID from client URL"
	errCaptureOrder       = "Error: main.go: Capture order"
	errRegisterSite       = "Error: main.go: Register site in database"
	errEncodeResponse     = "Error: main.go: Encode response"
	errCreateOrder        = "Error: main.go: Obtain orderID"
	errAuthGen            = "Error: main.go: Gen and register auth"
	errAuthEmail          = "Error: main.go: Send auth email"
	errAuthValidate       = "Error: main.go: Validate changes"
	errUpdateSite         = "Error: main.go: Updating site data"
)

type ConexData struct {
	Directory  string          `json:"directory"`
	Banner     string          `json:"banner"`
	Title      string          `json:"title"`
	Slogan     string          `json:"slogan"`
	EditorData json.RawMessage `json:"editor_data"`
}

func main() {
	var db *sql.DB
	var s3Client *s3.Client

	godotenv.Load()
	var (
		baseURL      = os.Getenv("BASE_URL")
		clientID     = os.Getenv("CLIENT_ID")
		clientSecret = os.Getenv("CLIENT_SECRET")
		returnURL    = os.Getenv("RETURN_URL")
		cancelURL    = os.Getenv("CANCEL_URL")
		port         = os.Getenv("PORT")
		amount       = os.Getenv("PRICE")
	)

	if baseURL == "" ||
		clientID == "" ||
		clientSecret == "" ||
		returnURL == "" ||
		cancelURL == "" ||
		port == "" {
		fatal(nil, errMissingCredentials)
	}

	var err error
	db, err = sql.Open("postgres", "host="+os.Getenv("DB_HOST")+
		" port="+os.Getenv("DB_PORT")+
		" user="+os.Getenv("DB_USER")+
		" password="+os.Getenv("DB_PASS")+
		" dbname="+os.Getenv("DB_NAME"))
	if err != nil {
		fatal(err, errDBConn)
	}

	if err := db.Ping(); err != nil {
		fatal(err, errDBPing)
	}

	msg(msgDBConn)

	var (
		bucketName     = os.Getenv("BUCKET_NAME")
		endpoint       = os.Getenv("BUCKET_ENDPOINT")
		accessKey      = os.Getenv("BUCKET_ACCESSKEY")
		secretKey      = os.Getenv("BUCKET_SECRETKEY")
		region         = os.Getenv("BUCKET_REGION")
		publicEndpoint = os.Getenv("BUCKET_PUBLIC_ENDPOINT")
		apiEndpoint    = os.Getenv("BUCKET_API_ENDPOINT")
		apiToken       = os.Getenv("BUCKET_API_TOKEN")
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		})),
	)
	if err != nil {
		fatal(err, errServerStart)
	}

	s3Client = s3.NewFromConfig(cfg)

	http.HandleFunc("/api/orders", CreateOrderHandler(db, amount))
	http.HandleFunc("/api/orders/", CaptureOrderHandler(db))
	http.HandleFunc("/api/update", UpdateSiteHandler(db))
	http.HandleFunc("/api/confirm", ConfirmChangesHandler(db))
	http.HandleFunc("/api/directory/", VerifyDirectoryHandler(db))
	http.HandleFunc("/api/fetch/", FetchSiteHandler(db))
	http.HandleFunc("/api/upload", UploadFileHandler(s3Client, endpoint, apiEndpoint, apiToken, bucketName, publicEndpoint))
	http.Handle("/", http.FileServer(http.Dir("./public")))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		msg(msgServerStart + ": " + port + "...")
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fatal(err, errServerStart)
		}
	}()

	<-stop

	if db != nil {
		msg(msgClosingDBConn)
		if err := db.Close(); err != nil {
			fatal(err, errClosingDBConn)
		}
	}
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
	log.Fatalf("%s: %v", notice, err)
}

func CreateOrderHandler(db *sql.DB, amount string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		var cart struct {
			Directory string `json:"directory"`
		}
		if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
			httpErrorAndLog(w, err, errReadBody, "Error decoding response")
			return
		}

		if len(cart.Directory) > 35 {
			http.Error(w, "Site already exists", http.StatusConflict)
			log.Printf("%s: %v", "Site title is too long", nil)
			return
		}

		if err := AvailableSite(db, cart.Directory); err != nil {
			http.Error(w, "Site already exists", http.StatusConflict)
			log.Printf("%s: %v", "Site already exists", err)
			return
		}

		orderID, err := CreateOrder(amount)
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
}

func CaptureOrderHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		errClientNotice := "Error capturing order"

		var cart ConexData
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

		if err := RegisterSitePayment(db, capture, cart); err != nil {
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
}

func UpdateSiteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		errClientNotice := "Error handling update request"

		var cart struct {
			Directory string `json:"directory"`
		}
		if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
			httpErrorAndLog(w, err, errReadBody, errClientNotice)
			return
		}

		code := GenerateCode()

		email, err := UpdateSiteAuth(db, cart.Directory, code)
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
}

func ConfirmChangesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		errClientNotice := "Error handling confirm changes request"

		var cart struct {
			Directory  string          `json:"directory"`
			Code       string          `json:"auth_code"`
			EditorData json.RawMessage `json:"editor_data"`
			Slogan     string          `json:"slogan"`
		}
		if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
			httpErrorAndLog(w, err, errReadBody, errClientNotice)
			return
		}

		pkey, err := ValidateSiteAuth(db, cart.Directory, cart.Code)
		if err != nil {
			httpErrorAndLog(w, err, errAuthValidate, errClientNotice)
			return
		}

		if err := UpdateSite(db, pkey, cart.EditorData, cart.Slogan); err != nil {
			httpErrorAndLog(w, err, errUpdateSite, errClientNotice)
			return
		}

		return
	}
}

func VerifyDirectoryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

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

		err := AvailableSite(db, folder)
		if err != nil {
			response.Exists = true
		} else {
			response.Exists = false
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
}

func UploadFileHandler(s3Client *s3.Client, endpoint string, apiEndpoint string,
	apiToken string, bucketName string, publicEndpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			httpErrorAndLog(w, err, "Unable to parse form", "Unable to parse form")
			return
		}
		directory := r.FormValue("directory")
		if directory == "" || len(directory) < 4 || len(directory) > 35 {
			err := fmt.Errorf("invalid directory length")
			httpErrorAndLog(w, err, "Unable to parse form", "Unable to parse form")
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			httpErrorAndLog(w, err, "Unable to get the file", "Unable to get the file")
			return
		}
		defer file.Close()

		fileContent, err := io.ReadAll(file)
		if err != nil {
			httpErrorAndLog(w, err, "Unable to read file", "Unable to read file")
			return
		}

		if len(fileContent) > maxUploadFileSize {
			httpErrorAndLog(w, err, "File too large", "File too large")
			return
		}

		if err := BucketSizeLimit(apiEndpoint, apiToken); err != nil {
			httpErrorAndLog(w, err, "Bucket limit", "Bucket limit")
			return
		}

		objectKey := fmt.Sprintf("%s/%s-%s", directory, time.Now().Format("2006-01-02-15-04-05"), fileHeader.Filename)
		url, err := UploadFile(s3Client, endpoint, bucketName, publicEndpoint, fileContent, objectKey)
		if err != nil {
			httpErrorAndLog(w, err, "Unable to upload file", "Unable to upload file")
			return
		}

		var response struct {
			Success int `json:"success"`
			File    struct {
				URL string `json:"url"`
			} `json:"file"`
		}

		response.Success = 1
		response.File.URL = url
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
}

func FetchSiteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		errClientNotice := "Error fetching site from db"

		path := strings.TrimPrefix(r.URL.Path, "/api/fetch/")
		parts := strings.Split(path, "/")
		folder := parts[0]
		if folder == "" {
			httpErrorAndLog(w, nil, "Error getting directory", errClientNotice)
			return
		}

		var siteData ConexData
		siteData, err := FetchSite(db, folder)
		if err != nil {
			httpErrorAndLog(w, err, "Error fetching site data", "Error fetching site data")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(siteData)
		return
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
