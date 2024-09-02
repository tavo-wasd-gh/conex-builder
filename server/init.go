package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	msgClosingDBConn = "Msg: init.go: Closing database connection"
	msgDBConn        = "Msg: init.go: Established database connection"
	errDBConn        = "Fatal: init.go: Connect to database"
	errDBPing        = "Fatal: init.go: Ping database"
	errClosingDBConn = "Fatal: init.go: Closing database connection"
)

var db *sql.DB

func initialize() {
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
		log.Fatalf("%s: %v", errDBConn, err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("%s: %v", errDBPing, err)
	}

	log.Println(msgDBConn)
}

func shutdown() {
	if db != nil {
		log.Println(msgClosingDBConn)
		if err := db.Close(); err != nil {
			log.Fatalf("%s: %v", errClosingDBConn, err)
		}
	}
}
