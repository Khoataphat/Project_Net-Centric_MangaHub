// cmd/seed/main.go
// =============================================================================
// MangaHub MangaDex Seeder
// Cào dữ liệu từ MangaDex API và nạp vào 3 bảng: mangas, chapters, pages.
//
// Chạy lệnh từ thư mục backend/:
//   go run ./cmd/seed/main.go
// hoặc chỉ định đường dẫn DB:
//   DB_PATH=./data/mangahub.db go run ./cmd/seed/main.go
// =============================================================================
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

// ─── Cấu hình chung ────────────────────────────────────────────────────────────

const (
	mangaDexBase   = "https://api.mangadex.org"
	uploadsBase    = "https://uploads.mangadex.org"
	mangaLimit     = 30  // Số truyện tối đa lấy từ API
	chapterLimit   = 10  // Số chương tối đa mỗi truyện
	requestDelay   = 400 * time.Millisecond // Rate limiting: ~2-3 req/s (an toàn)
	httpTimeout    = 15 * time.Second
)

// ─── Struct ánh xạ JSON từ MangaDex API ───────────────────────────────────────

// --- /manga endpoint ---

type MangaListResponse struct {
	Data []MangaData `json:"data"`
}

type MangaData struct {
	ID         string           `json:"id"`
	Attributes MangaAttributes  `json:"attributes"`
	Relations  []Relationship   `json:"relationships"`
}

type MangaAttributes struct {
	Title       map[string]string            `json:"title"`
	Description map[string]string            `json:"description"`
}

type Relationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes map[string]interface{} `json:"attributes"`
}

// --- /manga/{id}/feed endpoint ---

type ChapterFeedResponse struct {
	Data []ChapterData `json:"data"`
}

type ChapterData struct {
	ID         string              `json:"id"`
	Attributes ChapterAttributes  `json:"attributes"`
}

type ChapterAttributes struct {
	Chapter         *string `json:"chapter"`   // nullable string, ví dụ "1", "1.5"
	Title           *string `json:"title"`     // nullable
	TranslatedLang  string  `json:"translatedLanguage"`
}

// --- /at-home/server/{chapter_id} endpoint ---

type AtHomeResponse struct {
	BaseURL string          `json:"baseUrl"`
	Chapter AtHomeChapter  `json:"chapter"`
}

type AtHomeChapter struct {
	Hash string   `json:"hash"`
	Data []string `json:"data"` // filename list (chất lượng cao)
}

// ─── HTTP Client (singleton với timeout) ──────────────────────────────────────

var client = &http.Client{
	Timeout: httpTimeout,
}

// ─── Helper: thực hiện GET và decode JSON vào dst ────────────────────────────

func getJSON(url string, dst interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("request thất bại [%s]: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API trả về HTTP %d [%s]: %s", resp.StatusCode, url, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("lỗi parse JSON [%s]: %w", url, err)
	}
	return nil
}

// ─── Helper: lấy text từ map đa ngôn ngữ (ưu tiên en > ja-ro > lấy bất kỳ) ──

func pickText(m map[string]string) string {
	if v, ok := m["en"]; ok && v != "" {
		return v
	}
	for _, lang := range []string{"ja-ro", "zh", "ko"} {
		if v, ok := m[lang]; ok && v != "" {
			return v
		}
	}
	// fallback: lấy giá trị đầu tiên có trong map
	for _, v := range m {
		if v != "" {
			return v
		}
	}
	return ""
}

// ─── Helper: tìm cover filename từ danh sách relationships ───────────────────

func extractCoverFilename(rels []Relationship) string {
	for _, r := range rels {
		if r.Type == "cover_art" {
			if fn, ok := r.Attributes["fileName"].(string); ok {
				return fn
			}
		}
	}
	return ""
}

// ─── Helper: tìm tên tác giả từ danh sách relationships ──────────────────────

func extractAuthorName(rels []Relationship) string {
	for _, r := range rels {
		if r.Type == "author" {
			if name, ok := r.Attributes["name"].(string); ok {
				return name
			}
		}
	}
	return ""
}

// ─── Database: khởi tạo và migrate 3 bảng mới ────────────────────────────────

// addColumnIfNotExists kiểm tra và thêm cột vào bảng SQLite một cách an toàn.
func addColumnIfNotExists(db *sql.DB, table, column, definition string) {
	var count int
	query := fmt.Sprintf("SELECT count(*) FROM pragma_table_info('%s') WHERE name='%s'", table, column)
	if err := db.QueryRow(query).Scan(&count); err != nil {
		log.Printf("[DB] Không thể kiểm tra cột '%s.%s': %v", table, column, err)
		return
	}
	if count == 0 {
		alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
		log.Printf("[DB] Migration: %s", alterSQL)
		if _, err := db.Exec(alterSQL); err != nil {
			log.Printf("[DB] ALTER TABLE lỗi: %v", err)
		}
	}
}

func initDB(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("[DB] Không thể mở database: %v", err)
	}
	// SQLite: chỉ dùng 1 connection để tránh "database is locked"
	db.SetMaxOpenConns(1)

	// Bật journal_mode=DELETE để tương thích với Docker Volume trên Windows (tránh Disk I/O error)
	if _, err := db.Exec("PRAGMA journal_mode=DELETE;"); err != nil {
		log.Printf("[DB] PRAGMA journal_mode=DELETE thất bại: %v", err)
	}

	// ── Tạo bảng nếu chưa tồn tại ────────────────────────────────────────────
	createTables := []string{
		// Bảng users (giữ nguyên, không xóa dữ liệu cũ)
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// Bảng mangas (schema mới có mangadex_id; nếu đã tồn tại sẽ dùng ALTER bên dưới)
		`CREATE TABLE IF NOT EXISTS mangas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			author TEXT,
			description TEXT,
			thumbnail TEXT,
			mangadex_id TEXT UNIQUE
		);`,

		// Bảng chapters (MỚI)
		`CREATE TABLE IF NOT EXISTS chapters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			manga_id INTEGER NOT NULL,
			chapter_number REAL,
			title TEXT,
			mangadex_id TEXT UNIQUE,
			FOREIGN KEY (manga_id) REFERENCES mangas(id)
		);`,

		// Bảng pages (MỚI)
		`CREATE TABLE IF NOT EXISTS pages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chapter_id INTEGER NOT NULL,
			page_number INTEGER NOT NULL,
			image_url TEXT NOT NULL,
			FOREIGN KEY (chapter_id) REFERENCES chapters(id)
		);`,
	}

	for _, stmt := range createTables {
		if _, err := db.Exec(stmt); err != nil {
			log.Fatalf("[DB] Tạo bảng thất bại:\n%s\nLỗi: %v", stmt, err)
		}
	}

	// ── Migrations: thêm cột mới vào bảng cũ nếu cần ─────────────────────────
	// Bảng mangas: có thể đã tồn tại từ trước khi có mangadex_id
	addColumnIfNotExists(db, "mangas", "mangadex_id", "TEXT")
	// Không thể thêm UNIQUE constraint qua ALTER TABLE trong SQLite;
	// uniqueness được đảm bảo qua INSERT OR IGNORE logic trong code.

	log.Printf("[DB] Database đã sẵn sàng: %s", dbPath)
	return db
}

// ─── Ghi manga vào DB, trả về localID (hoặc -1 nếu lỗi) ─────────────────────

func upsertManga(db *sql.DB, mdxID, title, author, description, thumbnail string) int64 {
	// Dùng INSERT OR IGNORE để tránh trùng lặp theo mangadex_id
	res, err := db.Exec(
		`INSERT OR IGNORE INTO mangas (title, author, description, thumbnail, mangadex_id)
		 VALUES (?, ?, ?, ?, ?)`,
		title, author, description, thumbnail, mdxID,
	)
	if err != nil {
		log.Printf("  [WARN] Không thể insert manga '%s': %v", title, err)
		return -1
	}

	// Nếu IGNORE vì đã tồn tại, lấy ID hiện có
	lastID, _ := res.LastInsertId()
	if lastID == 0 {
		var existingID int64
		_ = db.QueryRow("SELECT id FROM mangas WHERE mangadex_id = ?", mdxID).Scan(&existingID)
		return existingID
	}
	return lastID
}

// ─── Ghi chapter vào DB, trả về localID (hoặc -1 nếu lỗi) ───────────────────

func upsertChapter(db *sql.DB, mangaID int64, mdxChapterID string, chapterNum float64, title string) int64 {
	res, err := db.Exec(
		`INSERT OR IGNORE INTO chapters (manga_id, chapter_number, title, mangadex_id)
		 VALUES (?, ?, ?, ?)`,
		mangaID, chapterNum, title, mdxChapterID,
	)
	if err != nil {
		log.Printf("  [WARN] Không thể insert chapter '%s': %v", mdxChapterID, err)
		return -1
	}

	lastID, _ := res.LastInsertId()
	if lastID == 0 {
		var existingID int64
		_ = db.QueryRow("SELECT id FROM chapters WHERE mangadex_id = ?", mdxChapterID).Scan(&existingID)
		return existingID
	}
	return lastID
}

// ─── Ghi batch pages vào DB ───────────────────────────────────────────────────

func insertPages(db *sql.DB, chapterID int64, imageURLs []string) int {
	// Kiểm tra xem chapter này đã có page chưa (để tránh duplicate khi chạy lại)
	var existingCount int
	_ = db.QueryRow("SELECT count(*) FROM pages WHERE chapter_id = ?", chapterID).Scan(&existingCount)
	if existingCount > 0 {
		return existingCount // Đã có, bỏ qua
	}

	// Dùng transaction để ghi nhanh hơn
	tx, err := db.Begin()
	if err != nil {
		log.Printf("  [WARN] Không thể bắt đầu transaction: %v", err)
		return 0
	}

	stmt, err := tx.Prepare(
		"INSERT INTO pages (chapter_id, page_number, image_url) VALUES (?, ?, ?)",
	)
	if err != nil {
		_ = tx.Rollback()
		log.Printf("  [WARN] Prepare statement lỗi: %v", err)
		return 0
	}
	defer stmt.Close()

	count := 0
	for i, url := range imageURLs {
		if _, err := stmt.Exec(chapterID, i+1, url); err != nil {
			log.Printf("  [WARN] Không thể insert page %d: %v", i+1, err)
			continue
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		log.Printf("  [WARN] Commit transaction lỗi: %v", err)
		return 0
	}
	return count
}

// ─── LUỒNG CHÍNH ─────────────────────────────────────────────────────────────

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("╔══════════════════════════════════════════════════╗")
	log.Println("║       MangaHub - MangaDex Seeder v1.0            ║")
	log.Println("╚══════════════════════════════════════════════════╝")

	// Lấy đường dẫn DB từ env hoặc mặc định
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/mangahub.db"
	}

	// Khởi tạo DB
	db := initDB(dbPath)
	defer db.Close()

	// ── BƯỚC 1: Lấy danh sách Manga ──────────────────────────────────────────
	log.Printf("\n[BƯỚC 1] Đang lấy %d truyện từ MangaDex...", mangaLimit)

	mangaURL := fmt.Sprintf(
		"%s/manga?limit=%d&includes[]=cover_art&includes[]=author&contentRating[]=safe&order[relevance]=desc",
		mangaDexBase, mangaLimit,
	)

	var mangaList MangaListResponse
	if err := getJSON(mangaURL, &mangaList); err != nil {
		log.Fatalf("[LỖI NGHIÊM TRỌNG] Không thể lấy danh sách manga: %v", err)
	}

	log.Printf("[BƯỚC 1] ✓ Lấy được %d truyện\n", len(mangaList.Data))
	time.Sleep(requestDelay)

	// Thống kê tổng kết
	totalMangas := 0
	totalChapters := 0
	totalPages := 0

	// ── BƯỚC 2 & 3: Lặp qua từng Manga ──────────────────────────────────────
	for i, manga := range mangaList.Data {
		title := pickText(manga.Attributes.Title)
		if title == "" {
			log.Printf("  [SKIP] Manga #%d (ID: %s) không có tiêu đề, bỏ qua.", i+1, manga.ID)
			continue
		}

		description := pickText(manga.Attributes.Description)
		author := extractAuthorName(manga.Relations)
		coverFilename := extractCoverFilename(manga.Relations)

		thumbnail := ""
		if coverFilename != "" {
			thumbnail = fmt.Sprintf("%s/covers/%s/%s", uploadsBase, manga.ID, coverFilename)
		}

		log.Printf("\n[%d/%d] Đang xử lý: %s", i+1, len(mangaList.Data), title)
		log.Printf("        ID: %s | Tác giả: %s", manga.ID, author)

		// Ghi manga vào DB
		mangaLocalID := upsertManga(db, manga.ID, title, author, description, thumbnail)
		if mangaLocalID < 0 {
			log.Printf("  [SKIP] Bỏ qua truyện '%s' do lỗi DB.", title)
			continue
		}
		totalMangas++
		log.Printf("        → Đã lưu manga (local ID: %d)", mangaLocalID)

		// Rate limit trước khi gọi tiếp
		time.Sleep(requestDelay)

		// ── BƯỚC 2: Lấy danh sách Chapters ──────────────────────────────────
		feedURL := fmt.Sprintf(
			"%s/manga/%s/feed?translatedLanguage[]=en&limit=%d&order[chapter]=asc&contentRating[]=safe",
			mangaDexBase, manga.ID, chapterLimit,
		)

		var feed ChapterFeedResponse
		if err := getJSON(feedURL, &feed); err != nil {
			log.Printf("  [WARN] Không lấy được chapter list cho '%s': %v", title, err)
			time.Sleep(requestDelay)
			continue
		}

		if len(feed.Data) == 0 {
			log.Printf("  [INFO] '%s' không có chương tiếng Anh, bỏ qua chapters.", title)
			continue
		}

		log.Printf("        → Tìm thấy %d chương tiếng Anh", len(feed.Data))
		chaptersAdded := 0

		for _, ch := range feed.Data {
			// Parse chapter number (string → float64)
			chapterNum := 0.0
			chapterNumStr := ""
			if ch.Attributes.Chapter != nil {
				chapterNumStr = *ch.Attributes.Chapter
				if parsed, err := strconv.ParseFloat(chapterNumStr, 64); err == nil {
					chapterNum = parsed
				}
			}

			chTitle := ""
			if ch.Attributes.Title != nil {
				chTitle = *ch.Attributes.Title
			}

			// Ghi chapter vào DB
			chLocalID := upsertChapter(db, mangaLocalID, ch.ID, chapterNum, chTitle)
			if chLocalID < 0 {
				continue
			}

			log.Printf("        [Ch %s] ID: %s → local ID: %d", chapterNumStr, ch.ID, chLocalID)

			time.Sleep(requestDelay)

			// ── BƯỚC 3: Lấy URLs trang cho Chapter này ───────────────────────
			atHomeURL := fmt.Sprintf("%s/at-home/server/%s", mangaDexBase, ch.ID)

			var atHome AtHomeResponse
			if err := getJSON(atHomeURL, &atHome); err != nil {
				log.Printf("          [WARN] Không lấy được at-home server cho ch %s: %v", chapterNumStr, err)
				time.Sleep(requestDelay)
				continue
			}

			if atHome.BaseURL == "" || atHome.Chapter.Hash == "" {
				log.Printf("          [WARN] at-home response thiếu dữ liệu cho ch %s, bỏ qua.", chapterNumStr)
				time.Sleep(requestDelay)
				continue
			}

			// Tạo URL đầy đủ cho mỗi trang
			imageURLs := make([]string, 0, len(atHome.Chapter.Data))
			for _, filename := range atHome.Chapter.Data {
				imageURL := fmt.Sprintf("%s/data/%s/%s",
					atHome.BaseURL,
					atHome.Chapter.Hash,
					filename,
				)
				imageURLs = append(imageURLs, imageURL)
			}

			// Ghi pages vào DB
			pagesAdded := insertPages(db, chLocalID, imageURLs)
			log.Printf("          → Đã lưu %d trang cho chương %s", pagesAdded, chapterNumStr)

			totalPages += pagesAdded
			chaptersAdded++

			time.Sleep(requestDelay)
		}

		totalChapters += chaptersAdded
		log.Printf("        ✓ Hoàn tất '%s': %d chương, ~%d trang", title, chaptersAdded, totalPages)
	}

	// ── KẾT QUẢ ──────────────────────────────────────────────────────────────
	log.Println("\n╔══════════════════════════════════════════════════╗")
	log.Println("║                 SEEDING HOÀN TẤT                ║")
	log.Println("╠══════════════════════════════════════════════════╣")
	log.Printf("║  %-18s %28d  ║\n", "Manga đã lưu:", totalMangas)
	log.Printf("║  %-18s %28d  ║\n", "Chapters đã lưu:", totalChapters)
	log.Printf("║  %-18s %28d  ║\n", "Pages đã lưu:", totalPages)
	log.Println("╚══════════════════════════════════════════════════╝")
	log.Printf("Database: %s\n", dbPath)
	log.Println("Bây giờ bạn có thể khởi động server và xem kết quả trên frontend!")
}
