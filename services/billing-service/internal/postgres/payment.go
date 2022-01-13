package postgres

import (
	"context"

	"billing-service/internal/billing"
	"billing-service/internal/postgres/database"

	"github.com/pkg/errors"
)

type PaymentRepository struct {
	db *database.Queries
}

func NewPaymentRepository(db *database.Queries) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (repository *PaymentRepository) Add(ctx context.Context, payment *billing.Payment) error {
	_, err := repository.db.CreatePayment(ctx, database.CreatePaymentParams{
		ID:        payment.ID,
		AccountID: payment.AccountID,
		Amount:    payment.Amount,
	})
	if err != nil {
		return errors.Wrap(err, "failed to add payment")
	}

	return nil
}
