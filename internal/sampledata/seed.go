package sampledata

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

var DBPath string

func Init() {
	dir := filepath.Join("/tmp", "qaynaq")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create sample data directory")
		return
	}

	DBPath = filepath.Join(dir, "sample.db")
	if _, err := os.Stat(DBPath); err == nil {
		return
	}

	db, err := sql.Open("sqlite", DBPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create sample database")
		return
	}
	defer db.Close()

	schema := `
CREATE TABLE customers (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE products (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  price REAL NOT NULL,
  category TEXT NOT NULL,
  stock INTEGER NOT NULL
);

CREATE TABLE orders (
  id INTEGER PRIMARY KEY,
  customer_id INTEGER NOT NULL,
  amount REAL NOT NULL,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY (customer_id) REFERENCES customers(id)
);
`
	if _, err := db.Exec(schema); err != nil {
		log.Error().Err(err).Msg("Failed to create sample schema")
		os.Remove(DBPath)
		return
	}

	customers := []struct {
		name, email, created string
	}{
		{"Alice Johnson", "alice@example.com", "2025-01-15"},
		{"Bob Smith", "bob@example.com", "2025-01-20"},
		{"Carol Davis", "carol@example.com", "2025-02-01"},
		{"Dan Wilson", "dan@example.com", "2025-02-10"},
		{"Eve Brown", "eve@example.com", "2025-02-15"},
		{"Frank Miller", "frank@example.com", "2025-03-01"},
		{"Grace Lee", "grace@example.com", "2025-03-05"},
		{"Hank Taylor", "hank@example.com", "2025-03-12"},
		{"Ivy Chen", "ivy@example.com", "2025-04-01"},
		{"Jack Moore", "jack@example.com", "2025-04-15"},
		{"Karen White", "karen@example.com", "2025-05-01"},
		{"Leo Harris", "leo@example.com", "2025-05-10"},
		{"Mia Clark", "mia@example.com", "2025-06-01"},
		{"Noah Lewis", "noah@example.com", "2025-06-15"},
		{"Olivia Hall", "olivia@example.com", "2025-07-01"},
		{"Paul Young", "paul@example.com", "2025-07-20"},
		{"Quinn King", "quinn@example.com", "2025-08-01"},
		{"Rita Scott", "rita@example.com", "2025-08-15"},
		{"Sam Green", "sam@example.com", "2025-09-01"},
		{"Tina Adams", "tina@example.com", "2025-09-10"},
	}

	for i, c := range customers {
		if _, err := db.Exec("INSERT INTO customers VALUES (?, ?, ?, ?)", i+1, c.name, c.email, c.created); err != nil {
			log.Error().Err(err).Msg("Failed to insert customer")
			return
		}
	}

	products := []struct {
		name     string
		price    float64
		category string
		stock    int
	}{
		{"Wireless Keyboard", 49.99, "Electronics", 150},
		{"USB-C Hub", 29.99, "Electronics", 200},
		{"Desk Lamp", 34.99, "Office", 80},
		{"Notebook Set", 12.99, "Office", 500},
		{"Coffee Mug", 9.99, "Kitchen", 300},
		{"Water Bottle", 14.99, "Kitchen", 250},
		{"Bluetooth Speaker", 59.99, "Electronics", 120},
		{"Mouse Pad", 8.99, "Office", 400},
		{"Webcam", 44.99, "Electronics", 90},
		{"Standing Desk Mat", 39.99, "Office", 60},
		{"Phone Stand", 19.99, "Electronics", 180},
		{"Pen Set", 7.99, "Office", 350},
		{"Travel Adapter", 24.99, "Electronics", 200},
		{"Tote Bag", 15.99, "Accessories", 220},
		{"Sunglasses", 22.99, "Accessories", 170},
	}

	for i, p := range products {
		if _, err := db.Exec("INSERT INTO products VALUES (?, ?, ?, ?, ?)", i+1, p.name, p.price, p.category, p.stock); err != nil {
			log.Error().Err(err).Msg("Failed to insert product")
			return
		}
	}

	orders := []struct {
		customerID int
		amount     float64
		status     string
		created    string
	}{
		{1, 49.99, "completed", "2025-01-20"},
		{1, 29.99, "completed", "2025-02-15"},
		{2, 34.99, "completed", "2025-02-01"},
		{3, 59.99, "completed", "2025-02-10"},
		{4, 12.99, "completed", "2025-02-20"},
		{5, 9.99, "completed", "2025-03-01"},
		{2, 44.99, "completed", "2025-03-05"},
		{6, 79.98, "completed", "2025-03-10"},
		{7, 24.99, "completed", "2025-03-15"},
		{8, 14.99, "completed", "2025-03-20"},
		{1, 39.99, "completed", "2025-04-01"},
		{3, 19.99, "completed", "2025-04-05"},
		{9, 104.97, "completed", "2025-04-10"},
		{10, 8.99, "completed", "2025-04-15"},
		{4, 49.99, "completed", "2025-04-20"},
		{5, 29.99, "completed", "2025-05-01"},
		{11, 34.99, "completed", "2025-05-05"},
		{12, 59.99, "completed", "2025-05-10"},
		{6, 22.99, "completed", "2025-05-15"},
		{7, 15.99, "completed", "2025-05-20"},
		{13, 49.99, "completed", "2025-06-01"},
		{14, 67.98, "completed", "2025-06-05"},
		{2, 12.99, "completed", "2025-06-10"},
		{8, 44.99, "completed", "2025-06-15"},
		{15, 9.99, "completed", "2025-06-20"},
		{16, 84.98, "completed", "2025-07-01"},
		{9, 29.99, "completed", "2025-07-05"},
		{1, 14.99, "completed", "2025-07-10"},
		{17, 39.99, "completed", "2025-07-15"},
		{10, 59.99, "completed", "2025-07-20"},
		{18, 24.99, "completed", "2025-08-01"},
		{3, 49.99, "completed", "2025-08-05"},
		{11, 19.99, "completed", "2025-08-10"},
		{19, 34.99, "completed", "2025-08-15"},
		{20, 8.99, "completed", "2025-08-20"},
		{4, 44.99, "completed", "2025-09-01"},
		{12, 15.99, "completed", "2025-09-05"},
		{5, 59.99, "completed", "2025-09-10"},
		{13, 29.99, "completed", "2025-09-15"},
		{6, 22.99, "completed", "2025-09-20"},
		{14, 49.99, "completed", "2025-10-01"},
		{7, 12.99, "completed", "2025-10-05"},
		{15, 39.99, "completed", "2025-10-10"},
		{8, 9.99, "completed", "2025-10-15"},
		{16, 34.99, "completed", "2025-10-20"},
		{17, 59.99, "completed", "2025-11-01"},
		{9, 24.99, "completed", "2025-11-05"},
		{18, 44.99, "completed", "2025-11-10"},
		{10, 14.99, "completed", "2025-11-15"},
		{19, 49.99, "completed", "2025-11-20"},
		{20, 29.99, "completed", "2025-12-01"},
		{1, 19.99, "completed", "2025-12-05"},
		{2, 67.98, "completed", "2025-12-10"},
		{11, 8.99, "completed", "2025-12-15"},
		{3, 34.99, "completed", "2025-12-20"},
		{12, 49.99, "completed", "2026-01-01"},
		{4, 29.99, "completed", "2026-01-05"},
		{13, 44.99, "completed", "2026-01-10"},
		{5, 59.99, "completed", "2026-01-15"},
		{14, 12.99, "completed", "2026-01-20"},
		{15, 39.99, "pending", "2026-02-01"},
		{6, 24.99, "pending", "2026-02-05"},
		{16, 9.99, "pending", "2026-02-10"},
		{7, 49.99, "pending", "2026-02-15"},
		{17, 34.99, "pending", "2026-02-20"},
		{8, 14.99, "shipped", "2026-03-01"},
		{18, 59.99, "shipped", "2026-03-05"},
		{9, 22.99, "shipped", "2026-03-10"},
		{19, 29.99, "shipped", "2026-03-15"},
		{10, 44.99, "shipped", "2026-03-20"},
		{20, 19.99, "processing", "2026-03-25"},
		{1, 84.98, "processing", "2026-03-26"},
		{2, 49.99, "processing", "2026-03-27"},
		{11, 34.99, "processing", "2026-03-28"},
		{3, 59.99, "processing", "2026-03-29"},
		{4, 12.99, "processing", "2026-03-30"},
		{12, 29.99, "processing", "2026-03-30"},
		{5, 44.99, "processing", "2026-03-30"},
		{13, 9.99, "processing", "2026-03-30"},
		{6, 39.99, "processing", "2026-03-30"},
		{14, 24.99, "processing", "2026-03-30"},
		{7, 49.99, "processing", "2026-03-30"},
		{15, 14.99, "processing", "2026-03-30"},
		{8, 34.99, "processing", "2026-03-30"},
		{16, 59.99, "processing", "2026-03-30"},
		{9, 19.99, "processing", "2026-03-30"},
		{17, 29.99, "processing", "2026-03-30"},
		{10, 44.99, "processing", "2026-03-30"},
		{18, 12.99, "processing", "2026-03-30"},
		{19, 49.99, "processing", "2026-03-30"},
		{20, 22.99, "processing", "2026-03-30"},
		{1, 34.99, "processing", "2026-03-30"},
		{2, 9.99, "processing", "2026-03-30"},
		{3, 39.99, "processing", "2026-03-30"},
		{4, 24.99, "processing", "2026-03-30"},
		{11, 59.99, "processing", "2026-03-30"},
		{5, 14.99, "processing", "2026-03-30"},
		{12, 49.99, "processing", "2026-03-30"},
		{6, 29.99, "processing", "2026-03-30"},
		{13, 44.99, "processing", "2026-03-30"},
		{7, 19.99, "processing", "2026-03-30"},
	}

	for i, o := range orders {
		if _, err := db.Exec("INSERT INTO orders VALUES (?, ?, ?, ?, ?)", i+1, o.customerID, o.amount, o.status, o.created); err != nil {
			log.Error().Err(err).Msg("Failed to insert order")
			return
		}
	}

	log.Info().Str("path", DBPath).Msg("Sample database created")
}
