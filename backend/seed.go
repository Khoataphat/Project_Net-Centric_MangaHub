// backend/seed.go
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	// Kết nối thẳng tới file db mà Người A đã tạo
	db, err := sql.Open("sqlite", "./data/mangahub.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Danh sách 5 bộ truyện mẫu
	mangas := []struct {
		Title       string
		Author      string
		Description string
		Thumbnail   string
	}{
		{"One Piece", "Eiichiro Oda", "Hành trình trở thành Vua Hải Tặc.", "https://upload.wikimedia.org/wikipedia/vi/9/90/One_Piece_Volume_101.jpg"},
		{"Naruto", "Masashi Kishimoto", "Cậu bé ninja với ước mơ trở thành Hokage.", "https://upload.wikimedia.org/wikipedia/vi/c/c7/Naruto_Volume_1_cover.jpg"},
		{"Dragon Ball", "Akira Toriyama", "Cuộc phiêu lưu tìm kiếm các viên ngọc rồng.", "https://upload.wikimedia.org/wikipedia/vi/d/d4/Dragon_Ball_v01_cover.jpg"},
		{"Doraemon", "Fujiko F. Fujio", "Chú mèo máy đến từ tương lai.", "https://upload.wikimedia.org/wikipedia/vi/b/b7/Doraemon_volume_1_cover.jpg"},
		{"Conan", "Gosho Aoyama", "Thám tử bị teo nhỏ đi phá án.", "https://upload.wikimedia.org/wikipedia/vi/6/6c/Conan_Volume_1_cover.jpg"},
	}

	fmt.Println("--- Đang chèn dữ liệu mẫu vào MangaHub ---")

	for _, m := range mangas {
		query := `INSERT INTO mangas (title, author, description, thumbnail) VALUES (?, ?, ?, ?)`
		_, err := db.Exec(query, m.Title, m.Author, m.Description, m.Thumbnail)
		if err != nil {
			fmt.Printf("Lỗi khi chèn %s: %v\n", m.Title, err)
		} else {
			fmt.Printf("Đã thêm thành công: %s\n", m.Title)
		}
	}

	fmt.Println("--- Hoàn tất! Bây giờ bạn có thể mở Dashboard để xem ---")
}
