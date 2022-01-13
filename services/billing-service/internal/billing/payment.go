package billing

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type Payment struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	Amount    float64
	CreatedAt time.Time
}

type PaymentRepository interface {
	Add(ctx context.Context, payment *Payment) error
}
