package http

import (
	"database/sql"
	"mangahub/internal/database"
	"mangahub/internal/models"
	"net/http"
	"github.com/gin-gonic/gin"
)

func GetMangas(c *gin.Context) {
	search := c.Query("q") // Lấy tham số tìm kiếm ?q=...
	query := "SELECT id, title, author, description, thumbnail FROM mangas"
	
	var rows *sql.Rows
	var err error

	if search != "" {
		query += " WHERE title LIKE ?"
		rows, err = database.DB.Query(query, "%"+search+"%")
	} else {
		rows, err = database.DB.Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var mangas []models.Manga
	for rows.Next() {
		var m models.Manga
		rows.Scan(&m.ID, &m.Title, &m.Author, &m.Description, &m.Thumbnail)
		mangas = append(mangas, m)
	}

	c.JSON(http.StatusOK, mangas)
}

func GetMangaByID(c *gin.Context) {
	id := c.Param("id")
	var m models.Manga

	query := "SELECT id, title, author, description, thumbnail FROM mangas WHERE id = ?"
	err := database.DB.QueryRow(query, id).Scan(&m.ID, &m.Title, &m.Author, &m.Description, &m.Thumbnail)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga không tồn tại"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, m)
}

func GetMangaChapters(c *gin.Context) {
	mangaID := c.Param("id")
	query := "SELECT id, manga_id, chapter_number, title FROM chapters WHERE manga_id = ? ORDER BY chapter_number ASC"
	rows, err := database.DB.Query(query, mangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var chapters []models.Chapter
	for rows.Next() {
		var ch models.Chapter
		rows.Scan(&ch.ID, &ch.MangaID, &ch.ChapterNumber, &ch.Title)
		chapters = append(chapters, ch)
	}
	c.JSON(http.StatusOK, chapters)
}

func GetChapterPages(c *gin.Context) {
	chapterID := c.Param("chapter_id")
	query := "SELECT id, chapter_id, page_number, image_url FROM pages WHERE chapter_id = ? ORDER BY page_number ASC"
	rows, err := database.DB.Query(query, chapterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var pages []models.Page
	for rows.Next() {
		var p models.Page
		rows.Scan(&p.ID, &p.ChapterID, &p.PageNumber, &p.ImageURL)
		pages = append(pages, p)
	}
	c.JSON(http.StatusOK, pages)
}