package messaging

import (
	"context"
	"encoding/json"

	"billing-service/internal/billing"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type CreatePayment struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"userId"`
	Amount float64   `json:"amount"`
}

type PaymentCreated struct {
	ID uuid.UUID `json:"id"`
}

func (p PaymentCreated) Name() string {
	return "Billing/PaymentCreated"
}

type PaymentFailed struct {
	ID uuid.UUID `json:"id"`
}

func (p PaymentFailed) Name() string {
	return "Billing/PaymentFailed"
}

type CreatePaymentProcessor struct {
	accounts        billing.AccountRepository
	payments        billing.PaymentRepository
	eventDispatcher EventDispatcher
}

func NewCreatePaymentProcessor(
	accounts billing.AccountRepository,
	payments billing.PaymentRepository,
	eventDispatcher EventDispatcher,
) *CreatePaymentProcessor {
	return &CreatePaymentProcessor{
		accounts:        accounts,
		payments:        payments,
		eventDispatcher: eventDispatcher,
	}
}

func (processor *CreatePaymentProcessor) Process(ctx context.Context, message []byte) error {
	var command CreatePayment
	err := json.Unmarshal(message, &command)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal CreatePayment command")
	}

	account, err := processor.accounts.FindByID(ctx, command.UserID)
	if err != nil {
		return errors.WithMessagef(err, `failed to find user account by id "%s"`, command.UserID)
	}

	account.Amount -= command.Amount
	if account.Amount < 0 {
		err = processor.eventDispatcher.Dispatch(ctx, PaymentFailed{ID: command.ID})
		if err != nil {
			return errors.WithMessage(err, "failed to dispatch PaymentFailed event")
		}

		return nil
	}

	err = processor.payments.Add(ctx, &billing.Payment{
		ID:        command.ID,
		AccountID: command.UserID,
		Amount:    command.Amount,
	})
	if err != nil {
		return errors.WithMessagef(err, `failed to add payment "%s" for user "%s"`, command.ID, command.UserID)
	}

	err = processor.accounts.Save(ctx, account)
	if err != nil {
		return errors.WithMessagef(err, `failed to save user account "%s"`, account.ID)
	}

	err = processor.eventDispatcher.Dispatch(ctx, PaymentCreated{ID: command.ID})
	if err != nil {
		return errors.WithMessage(err, "failed to dispatch PaymentCreated event")
	}

	return nil
}
