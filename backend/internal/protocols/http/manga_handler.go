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