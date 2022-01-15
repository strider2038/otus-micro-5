package ordering

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type OrderStatus string

const (
	StatusPending   = "pending"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

type Order struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"-"`
	PaymentID uuid.UUID   `json:"-"`
	Price     float64     `json:"price"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

func NewOrder(userID uuid.UUID, price float64) *Order {
	return &Order{
		UserID:    userID,
		PaymentID: uuid.Must(uuid.NewV4()),
		Price:     price,
		Status:    StatusPending,
	}
}

type OrderRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	FindByUser(ctx context.Context, userID uuid.UUID) ([]*Order, error)
	FindByPayment(ctx context.Context, paymentID uuid.UUID) (*Order, error)
	Save(ctx context.Context, order *Order) error
}
