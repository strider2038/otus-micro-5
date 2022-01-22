package postgres

import (
	"context"

	"order-service/internal/ordering"
	"order-service/internal/postgres/database"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type OrderRepository struct {
	db *database.Queries
}

func NewOrderRepository(db *database.Queries) *OrderRepository {
	return &OrderRepository{db: db}
}

func (repository *OrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*ordering.Order, error) {
	order, err := repository.db.FindOrder(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find order")
	}

	return orderFromDB(order), nil
}

func (repository *OrderRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]*ordering.Order, error) {
	dbOrders, err := repository.db.FindUserOrders(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find orders")
	}

	orders := make([]*ordering.Order, len(dbOrders))
	for i := range dbOrders {
		orders[i] = orderFromDB(dbOrders[i])
	}

	return orders, nil
}

func (repository *OrderRepository) FindByPayment(ctx context.Context, paymentID uuid.UUID) (*ordering.Order, error) {
	order, err := repository.db.FindOrderByPayment(ctx, paymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find order by payment")
	}

	return orderFromDB(order), nil
}

func (repository *OrderRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := repository.db.CountOrdersByUser(ctx, userID)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count orders by user")
	}

	return int(count), nil
}

func (repository *OrderRepository) Save(ctx context.Context, order *ordering.Order) error {
	if order.ID.IsNil() {
		order.ID = uuid.Must(uuid.NewV4())

		dbOrder, err := repository.db.CreateOrder(ctx, database.CreateOrderParams{
			ID:        order.ID,
			Price:     order.Price,
			UserID:    order.UserID,
			PaymentID: order.PaymentID,
		})
		if err != nil {
			return errors.Wrap(err, "failed to create order")
		}

		order.Status = ordering.OrderStatus(dbOrder.Status)
		order.CreatedAt = dbOrder.CreatedAt
		order.UpdatedAt = dbOrder.UpdatedAt

		return nil
	}

	dbOrder, err := repository.db.UpdateOrder(ctx, database.UpdateOrderParams{
		ID:     order.ID,
		Status: string(order.Status),
	})
	if err != nil {
		return errors.Wrap(err, "failed to update order")
	}

	order.Status = ordering.OrderStatus(dbOrder.Status)
	order.CreatedAt = dbOrder.CreatedAt
	order.UpdatedAt = dbOrder.UpdatedAt

	return nil
}

func orderFromDB(order database.Order) *ordering.Order {
	return &ordering.Order{
		ID:        order.ID,
		UserID:    order.UserID,
		PaymentID: order.PaymentID,
		Price:     order.Price,
		Status:    ordering.OrderStatus(order.Status),
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}
