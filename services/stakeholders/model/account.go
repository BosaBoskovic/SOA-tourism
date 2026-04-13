package model

import "time"

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=100"`
	Email    string `json:"email" binding:"required,email,max=120"`
	Role     string `json:"role" binding:"required"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"usernameOrEmail" binding:"required,min=3,max=120"`
	Password        string `json:"password" binding:"required,min=8,max=100"`
}

type AccountResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsBlocked bool      `json:"isBlocked"`
	CreatedAt time.Time `json:"createdAt"`
}

type Account struct {
	Username     string
	Email        string
	Role         string
	IsBlocked    bool
	PasswordHash string
}
