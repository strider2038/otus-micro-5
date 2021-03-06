// Code generated by sqlc. DO NOT EDIT.

package database

import (
	"time"

	"github.com/gofrs/uuid"
)

type Account struct {
	ID        uuid.UUID
	Amount    float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Payment struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	Amount    float64
	CreatedAt time.Time
}
