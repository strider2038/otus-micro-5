-- name: FindAccount :one
SELECT id, amount, created_at, updated_at
FROM "account"
WHERE id = $1
LIMIT 1;

-- name: CreateAccount :one
INSERT INTO "account" (id)
VALUES ($1)
RETURNING id, amount, created_at, updated_at;

-- name: UpdateAccount :one
UPDATE "account"
SET amount = $2, updated_at = now()
WHERE id = $1
RETURNING id, amount, created_at, updated_at;

-- name: CreatePayment :one
INSERT INTO "payment" (id, account_id, amount)
VALUES ($1, $2, $3)
RETURNING id, account_id, amount, created_at;
