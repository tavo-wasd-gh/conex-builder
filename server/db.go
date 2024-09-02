package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func RegisterSitePayment(capture Capture, directory string, editorData json.RawMessage) error {
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
			return fmt.Errorf("Error: Could not register site to database: %v", err)
		}
	} else {
		if err := db.QueryRow(
			`UPDATE sites SET due = due + INTERVAL '1 year'
			WHERE id = $1
			RETURNING id`,
			pkey).Scan(&pkey); err != nil {
			return fmt.Errorf("Error: Could not update due date: %v", err)
		}
	}

	if _, err := db.Exec(
		`INSERT INTO payments (capture, site, amount, currency, date, status)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		id, pkey, amount, currency, date, pstatus); err != nil {
		return fmt.Errorf("Error: Could not register payment to database: %v", err)
	}

	return nil
}

func UpdateSite(by string, pkey int, editorData json.RawMessage) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Error: Could not start transaction: %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var prev json.RawMessage
	if err := tx.QueryRow(
		`SELECT raw FROM sites WHERE id = $1`,
		pkey).Scan(&prev); err != nil {
		return fmt.Errorf("Error: Could not retrieve old value: %v", err)
	}

	if _, err = tx.Exec(
		`UPDATE sites SET raw = $1 WHERE id = $2`,
		editorData, pkey); err != nil {
		return fmt.Errorf("Error: Could not update raw column: %v", err)
	}

	if _, err = tx.Exec(
		`INSERT INTO changes (by, site, payment, col, prev, next, date)
		VALUES ($1, $2, NULL, 'raw', $3, $4, CURRENT_DATE);`,
		by, pkey, prev, editorData); err != nil {
		return fmt.Errorf("Error: Could not register change to database: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Error: Could not commit transaction: %v", err)
	}

	return nil
}
