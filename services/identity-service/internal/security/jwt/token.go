package jwt

import "github.com/golang-jwt/jwt/v4"

const KeyID = "auth"

type Claims struct {
	jwt.RegisteredClaims
	UserID int64  `json:"userId"`
	Email  string `json:"email"`
}
