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

	// Đảm bảo SQLite chỉ dùng 1 kết nối duy nhất để tránh lỗi "database is locked" khi ghi đồng thời (AC2)
	DB.SetMaxOpenConns(1)

	// 1. Tạo bảng users nếu chưa tồn tại
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

	// 2. Tạo bảng mangas nếu chưa tồn tại
	createMangaTableSQL := `
	CREATE TABLE IF NOT EXISTS mangas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT,
		description TEXT,
		thumbnail TEXT
	);`

	_, err = DB.Exec(createMangaTableSQL)
	if err != nil {
		log.Fatal("Lỗi tạo bảng mangas:", err)
	}

	// 3. Tạo bảng user_progress nếu chưa tồn tại (Mới - Tuần 4)
	createProgressTableSQL := `
	CREATE TABLE IF NOT EXISTS user_progress (
		user_id INTEGER NOT NULL,
		manga_id INTEGER NOT NULL,
		last_chapter INTEGER DEFAULT 1,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, manga_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (manga_id) REFERENCES mangas(id)
	);`

	_, err = DB.Exec(createProgressTableSQL)
	if err != nil {
		log.Fatal("Lỗi tạo bảng user_progress:", err)
	}

	log.Println("Database SQLite đã sẵn sàng tại:", filepath)
}
