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

	// Ép journal mode về DELETE để tương thích với Docker Volume trên Windows (tránh Disk I/O error)
	_, _ = DB.Exec("PRAGMA journal_mode=DELETE;")

	// 1. Tạo bảng users nếu chưa tồn tại
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'user',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Lỗi tạo bảng users:", err)
	}

	// Migration: Kiểm tra xem cột 'role' đã tồn tại chưa, nếu chưa thì thêm vào (Trường hợp bảng đã tồn tại từ trước)
	var count int
	err = DB.QueryRow("SELECT count(*) FROM pragma_table_info('users') WHERE name='role'").Scan(&count)
	if err == nil && count == 0 {
		_, _ = DB.Exec("ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user'")
	}

	// 2. Tạo bảng mangas nếu chưa tồn tại
	createMangaTableSQL := `
	CREATE TABLE IF NOT EXISTS mangas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT,
		description TEXT,
		thumbnail TEXT,
		mangadex_id TEXT UNIQUE
	);`

	_, err = DB.Exec(createMangaTableSQL)
	if err != nil {
		log.Fatal("Lỗi tạo bảng mangas:", err)
	}

	// Migration: Thêm mangadex_id nếu chưa có
	err = DB.QueryRow("SELECT count(*) FROM pragma_table_info('mangas') WHERE name='mangadex_id'").Scan(&count)
	if err == nil && count == 0 {
		_, _ = DB.Exec("ALTER TABLE mangas ADD COLUMN mangadex_id TEXT UNIQUE")
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

	// 4. Tạo bảng chapters và pages nếu chưa có
	createChaptersTableSQL := `
	CREATE TABLE IF NOT EXISTS chapters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		manga_id INTEGER NOT NULL,
		chapter_number REAL,
		title TEXT,
		mangadex_id TEXT UNIQUE,
		FOREIGN KEY (manga_id) REFERENCES mangas(id)
	);`
	_, _ = DB.Exec(createChaptersTableSQL)

	createPagesTableSQL := `
	CREATE TABLE IF NOT EXISTS pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chapter_id INTEGER NOT NULL,
		page_number INTEGER NOT NULL,
		image_url TEXT NOT NULL,
		FOREIGN KEY (chapter_id) REFERENCES chapters(id)
	);`
	_, _ = DB.Exec(createPagesTableSQL)

	log.Println("Database SQLite đã sẵn sàng tại:", filepath)
}
