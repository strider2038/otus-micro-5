package users

import (
	"context"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID        int64  `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"-"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type Repository interface {
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	CountByEmail(ctx context.Context, email string) (int64, error)
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
}
