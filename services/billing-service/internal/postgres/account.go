package postgres

import (
	"context"

	"billing-service/internal/billing"
	"billing-service/internal/postgres/database"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type AccountRepository struct {
	db *database.Queries
}

func NewAccountRepository(db *database.Queries) *AccountRepository {
	return &AccountRepository{db: db}
}

func (repository *AccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*billing.Account, error) {
	account, err := repository.db.FindAccount(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, billing.ErrAccountNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to find account")
	}

	return accountFromDB(account), nil
}

func (repository *AccountRepository) Create(ctx context.Context, id uuid.UUID) (*billing.Account, error) {
	account, err := repository.db.CreateAccount(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create account")
	}

	return accountFromDB(account), nil
}

func (repository *AccountRepository) Save(ctx context.Context, account *billing.Account) error {
	a, err := repository.db.UpdateAccount(ctx, database.UpdateAccountParams{
		ID:     account.ID,
		Amount: account.Amount,
	})
	if err != nil {
		return errors.Wrap(err, "failed to save account")
	}

	account.Amount = a.Amount
	account.UpdatedAt = a.UpdatedAt

	return nil
}

func accountFromDB(account database.Account) *billing.Account {
	return &billing.Account{
		ID:        account.ID,
		Amount:    account.Amount,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}
}
