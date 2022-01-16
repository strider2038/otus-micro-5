/*
 * Billing service
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api

import (
	"context"
	"net/http"

	"billing-service/internal/billing"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/strider2038/pkg/persistence"
)

// BillingApiService is a service that implents the logic for the BillingApiServicer
// This service should implement the business logic for every endpoint for the BillingApi API.
// Include any external packages or services that will be required by this service.
type BillingApiService struct {
	accounts           billing.AccountRepository
	transactionManager persistence.TransactionManager
}

// NewBillingApiService creates a default api service
func NewBillingApiService(
	accounts billing.AccountRepository,
	transactionManager persistence.TransactionManager,
) BillingApiServicer {
	return &BillingApiService{accounts: accounts, transactionManager: transactionManager}
}

// GetBillingAccount -
func (s *BillingApiService) GetBillingAccount(ctx context.Context, id uuid.UUID) (ImplResponse, error) {
	account, err := s.accounts.FindByID(ctx, id)
	if err != nil {
		return Response(http.StatusInternalServerError, nil), err
	}

	return Response(http.StatusOK, account), nil
}

// DepositMoney -
func (s *BillingApiService) DepositMoney(ctx context.Context, form AccountUpdateForm) (ImplResponse, error) {
	err := s.transactionManager.DoTransactionally(ctx, func(ctx context.Context) error {
		account, err := s.accounts.FindByIDForUpdate(ctx, form.ID)
		if err != nil {
			return errors.WithMessagef(err, "failed to find account %s", form.ID)
		}

		account.Amount += form.Amount

		err = s.accounts.Save(ctx, account)
		if err != nil {
			return errors.WithMessagef(err, "failed to save account %s", form.ID)
		}

		return nil
	})
	if err != nil {
		return Response(http.StatusInternalServerError, nil), err
	}

	return Response(http.StatusNoContent, nil), nil
}

// WithdrawMoney -
func (s *BillingApiService) WithdrawMoney(ctx context.Context, form AccountUpdateForm) (ImplResponse, error) {
	err := s.transactionManager.DoTransactionally(ctx, func(ctx context.Context) error {
		account, err := s.accounts.FindByIDForUpdate(ctx, form.ID)
		if err != nil {
			return errors.WithMessagef(err, "failed to find account %s", form.ID)
		}

		account.Amount -= form.Amount
		if account.Amount < 0 {
			return billing.ErrNotEnoughMoney
		}

		err = s.accounts.Save(ctx, account)
		if err != nil {
			return errors.WithMessagef(err, "failed to save account %s", form.ID)
		}

		return nil
	})
	if errors.Is(err, billing.ErrNotEnoughMoney) {
		return Response(http.StatusUnprocessableEntity, Error{
			Code:    http.StatusUnprocessableEntity,
			Message: "Not enough money on the account",
		}), nil
	}
	if err != nil {
		return Response(http.StatusInternalServerError, nil), err
	}

	return Response(http.StatusNoContent, nil), nil
}
