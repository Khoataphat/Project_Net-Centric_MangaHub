package database

import (
	"database/sql"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

var DB *sql.DB

func InitDB(filepath string) {
	var err error
	DB, err = sql.Open("sqlite", filepath)
	if err != nil {
		log.Fatal("Không thể kết nối database:", err)
	}

	// Tạo bảng users nếu chưa tồn tại
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Lỗi tạo bảng users:", err)
	}

	log.Println("Database SQLite đã sẵn sàng tại:", filepath)
}
