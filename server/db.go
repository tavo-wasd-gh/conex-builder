package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	errDBRegisterSite    = "db.go (sites): Register site"
	errDBUpdateDue       = "db.go (sites): Update due date"
	errDBRegisterPayment = "db.go (payments): Register payment"
	errDBUpdateRaw       = "db.go (sites): Update raw json"
	errDBChangesRaw      = "db.go (changes): Register raw json change"
	errDBUpdateSiteAuth  = "db.go (sites): Auth"
)

func AvailableSite(db *sql.DB, folder string) error {
	if len(folder) <= 3 {
		return fmt.Errorf("folder name must be longer than 3 characters")
	}

	var exists bool
	if err := db.QueryRow(`
		SELECT EXISTS(SELECT * FROM sites WHERE folder = $1)
		`, folder).Scan(&exists); err != nil {
		return fmt.Errorf("error checking if folder exists: %v", err)
	}

	if exists {
		return fmt.Errorf("folder %s already exists", folder)
	}

	return nil
}

func RegisterSitePayment(db *sql.DB, capture Capture, cart ConexData) error {
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
		// conex_data
		directory  string
		title      string
		slogan     string
		banner     string
		editorData json.RawMessage
	)

	captureData := capture.PurchaseUnits[0].Payments.Captures[0]

	id = captureData.ID
	amount = captureData.Amount.Value
	currency = captureData.Amount.CurrencyCode
	pstatus = captureData.Status
	date = captureData.CreateTime
	wstatus = "down"
	due = date.AddDate(1, 0, 0)
	name = capture.Payer.Name.GivenName
	surname = capture.Payer.Name.Surname
	email = capture.Payer.EmailAddress
	phone = capture.Payer.Phone.PhoneNumber.NationalNumber
	country = capture.Payer.Address.CountryCode

	directory = cart.Directory
	title = cart.Title
	slogan = cart.Slogan
	banner = cart.Banner
	editorData = cart.EditorData

	var pkey int
	newSite := db.QueryRow(`
		SELECT id FROM sites WHERE folder = $1
		`, directory).Scan(&pkey)

	if newSite == sql.ErrNoRows {
		if err := db.QueryRow(`
			INSERT INTO sites
			(folder, status, due, name, sur, email, phone, code, title, slogan, banner, raw)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id
			`, directory, wstatus, due,
			name, surname, email, phone, country, title, slogan,
			banner, editorData).Scan(&pkey); err != nil {
			return fmt.Errorf("%s: %v", errDBRegisterSite, err)
		}
	} else {
		if err := db.QueryRow(`
			UPDATE sites SET due = due + INTERVAL '1 year'
			WHERE id = $1
			RETURNING id
			`, pkey).Scan(&pkey); err != nil {
			return fmt.Errorf("%s: %v", errDBUpdateDue, err)
		}
	}

	if _, err := db.Exec(`
		INSERT INTO payments
		(capture, site, amount, currency, date, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		`, id, pkey, amount, currency, date, pstatus); err != nil {
		return fmt.Errorf("%s: %v", errDBRegisterPayment, err)
	}

	return nil
}

func UpdateSite(db *sql.DB, pkey int, editorData json.RawMessage, slogan string) error {
	if _, err := db.Exec(`
		UPDATE sites
		SET raw = $1, slogan = $2, status = 'diff'
		WHERE id = $3
		`, editorData, slogan, pkey); err != nil {
		return fmt.Errorf("%s: %v", errDBUpdateRaw, err)
	}

	return nil
}

func UpdateSiteAuth(db *sql.DB, folder string, code string) (string, error) {
	var valid sql.NullTime
	if err := db.QueryRow(`
		SELECT valid
		FROM sites
		WHERE folder = $1
	`, folder).Scan(&valid); err != nil {
		return "", fmt.Errorf("error fetching valid timestamp: %v", err)
	}

	if valid.Valid && valid.Time.After(time.Now().Add(4*time.Minute)) {
		return "", fmt.Errorf("valid timestamp is still active, cannot update")
	}

	newValid := time.Now().Add(5 * time.Minute)
	var email string
	if err := db.QueryRow(`
		UPDATE sites
		SET auth = $1, valid = $2
		WHERE folder = $3
		RETURNING email;
		`, code, newValid, folder).Scan(&email); err != nil {
		return "", fmt.Errorf("%s: %v", errDBUpdateSiteAuth, err)
	}

	return email, nil
}

func ValidateSiteAuth(db *sql.DB, folder string, code string) (int, error) {
	var dbCode string
	var validTime time.Time

	err := db.QueryRow(`
		SELECT auth, valid FROM sites
		WHERE folder = $1;
		`, folder).Scan(&dbCode, &validTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("%s: %v", "No such directory", err)
		}
		return 0, fmt.Errorf("%s: %v", "Failed to query DB", err)
	}

	if code != dbCode {
		return 0, fmt.Errorf("%s", "Incorrect code")
	}

	if time.Now().After(validTime) {
		return 0, fmt.Errorf("%s", "Auth expired")
	}

	var pkey int
	if err := db.QueryRow(`
		UPDATE sites
		SET valid = $1
		WHERE folder = $2
		RETURNING id;
		`, time.Now(), folder).Scan(&pkey); err != nil {
		return 0, fmt.Errorf("%s: %v", "Void used code", err)
	}

	return pkey, nil
}

func FetchSite(db *sql.DB, folder string) (ConexData, error) {
	var siteData ConexData

	query := `
		SELECT folder, banner, title, slogan, raw 
		FROM sites 
		WHERE folder = $1
	`

	var rawData []byte

	err := db.QueryRow(query, folder).Scan(
		&siteData.Directory,
		&siteData.Banner,
		&siteData.Title,
		&siteData.Slogan,
		&rawData,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return siteData, fmt.Errorf("site not found: %v", folder)
		}
		return siteData, fmt.Errorf("error fetching site: %v", err)
	}

	err = json.Unmarshal(rawData, &siteData.EditorData)
	if err != nil {
		return siteData, fmt.Errorf("error unmarshaling editor data: %v", err)
	}

	return siteData, nil
}
