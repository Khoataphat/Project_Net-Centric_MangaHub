package models

import "time"

// User
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"password,omitempty" binding:"required"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Yêu cầu đăng nhập từ client
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

//Struct Manga
type Manga struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
}

type Chapter struct {
	ID            int     `json:"id"`
	MangaID       int     `json:"manga_id"`
	ChapterNumber float64 `json:"chapter_number"`
	Title         string  `json:"title"`
}

type Page struct {
	ID         int    `json:"id"`
	ChapterID  int    `json:"chapter_id"`
	PageNumber int    `json:"page_number"`
	ImageURL   string `json:"image_url"`
}
