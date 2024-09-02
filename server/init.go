package main

import (
	"database/sql"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	msgClosingDBConn      = "Msg: init.go: Closing database connection"
	msgDBConn             = "Msg: init.go: Established database connection"
	errDBConn             = "Fatal: init.go: Connect to database"
	errDBPing             = "Fatal: init.go: Ping database"
	errClosingDBConn      = "Fatal: init.go: Closing database connection"
	errMissingCredentials = "Fatal: init.go: Credentials"
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
}

func shutdown() {
	if db != nil {
		msg(msgClosingDBConn)
		if err := db.Close(); err != nil {
			fatal(err, errClosingDBConn)
		}
	}
}
