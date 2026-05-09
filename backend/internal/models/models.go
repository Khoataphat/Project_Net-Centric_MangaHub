package models

import "time"

// User
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"password,omitempty" binding:"required"`
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
