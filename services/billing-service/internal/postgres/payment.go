package postgres

import (
	"context"

	"billing-service/internal/billing"
	"billing-service/internal/postgres/database"

	"github.com/pkg/errors"
	postgres "github.com/strider2038/pkg/persistence/pgx"
)

type PaymentRepository struct {
	conn postgres.Connection
}

func NewPaymentRepository(conn postgres.Connection) *PaymentRepository {
	return &PaymentRepository{conn: conn}
}

func (repository *PaymentRepository) Add(ctx context.Context, payment *billing.Payment) error {
	_, err := database.New(repository.conn.Scope(ctx)).CreatePayment(ctx, database.CreatePaymentParams{
		ID:        payment.ID,
		AccountID: payment.AccountID,
		Amount:    payment.Amount,
	})
	if err != nil {
		return errors.Wrap(err, "failed to add payment")
	}

	return nil
}
