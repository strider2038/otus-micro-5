package messaging

import "github.com/gofrs/uuid"

type OrderSucceeded struct {
	UserID uuid.UUID `json:"userId"`
	Price  float64   `json:"price"`
}

func (o OrderSucceeded) Name() string {
	return "Ordering/OrderSucceeded"
}

type OrderFailed struct {
	UserID uuid.UUID `json:"userId"`
	Price  float64   `json:"price"`
	Reason string    `json:"reason"`
}

func (o OrderFailed) Name() string {
	return "Ordering/OrderFailed"
}
