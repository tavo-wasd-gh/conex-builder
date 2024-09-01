package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
		log.Fatalf("Error 001: Can't connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Error 001: Can't connect to database: %v", err)
	}

	log.Println("Established database connection")
}

func shutdown() {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Fatalf("Error: Can't close database connection: %v", err)
		}
	}
}
