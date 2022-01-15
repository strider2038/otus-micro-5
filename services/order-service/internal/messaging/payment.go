package messaging

import (
	"context"
	"encoding/json"

	"order-service/internal/ordering"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type CreatePayment struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"userId"`
	Amount float64   `json:"amount"`
}

func (c CreatePayment) Name() string {
	return "Billing/CreatePayment"
}

type PaymentCreated struct {
	ID uuid.UUID `json:"id"`
}

type PaymentFailed struct {
	ID     uuid.UUID `json:"id"`
	Reason string    `json:"reason"`
}

type PaymentCreatedProcessor struct {
	orders     ordering.OrderRepository
	dispatcher Dispatcher
}

func NewPaymentCreatedProcessor(orders ordering.OrderRepository, dispatcher Dispatcher) *PaymentCreatedProcessor {
	return &PaymentCreatedProcessor{orders: orders, dispatcher: dispatcher}
}

func (processor *PaymentCreatedProcessor) Process(ctx context.Context, message []byte) error {
	var payment PaymentCreated
	err := json.Unmarshal(message, &payment)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal PaymentCreated event")
	}

	order, err := processor.orders.FindByPayment(ctx, payment.ID)
	if err != nil {
		return errors.WithMessagef(err, "failed to find order by payment %s", payment.ID)
	}

	order.Status = ordering.StatusSucceeded
	err = processor.orders.Save(ctx, order)
	if err != nil {
		return errors.WithMessagef(err, "failed to save order %s of user %s", order.ID, order.UserID)
	}

	err = processor.dispatcher.Dispatch(ctx, &OrderSucceeded{
		UserID: order.UserID,
		Price:  order.Price,
	})
	if err != nil {
		return errors.WithMessagef(err, "failed to dispatch OrderSucceeded event for user %s", order.UserID)
	}

	return nil
}

type PaymentFailedProcessor struct {
	orders     ordering.OrderRepository
	dispatcher Dispatcher
}

func NewPaymentFailedProcessor(orders ordering.OrderRepository, dispatcher Dispatcher) *PaymentFailedProcessor {
	return &PaymentFailedProcessor{orders: orders, dispatcher: dispatcher}
}

func (processor *PaymentFailedProcessor) Process(ctx context.Context, message []byte) error {
	var payment PaymentFailed
	err := json.Unmarshal(message, &payment)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal PaymentFailed event")
	}

	order, err := processor.orders.FindByPayment(ctx, payment.ID)
	if err != nil {
		return errors.WithMessagef(err, "failed to find order by payment %s", payment.ID)
	}

	order.Status = ordering.StatusFailed
	err = processor.orders.Save(ctx, order)
	if err != nil {
		return errors.WithMessagef(err, "failed to save order %s of user %s", order.ID, order.UserID)
	}

	err = processor.dispatcher.Dispatch(ctx, &OrderFailed{
		UserID: order.UserID,
		Price:  order.Price,
		Reason: payment.Reason,
	})
	if err != nil {
		return errors.WithMessagef(err, "failed to dispatch OrderFailed event for user %s", order.UserID)
	}

	return nil
}
