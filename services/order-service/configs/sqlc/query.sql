-- name: FindUserOrders :many
SELECT id, user_id, payment_id, price, status, created_at, updated_at
FROM "order"
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: FindOrder :one
SELECT id, user_id, payment_id, price, status, created_at, updated_at
FROM "order"
WHERE id = $1
LIMIT 1;

-- name: FindOrderByPayment :one
SELECT id, user_id, payment_id, price, status, created_at, updated_at
FROM "order"
WHERE payment_id = $1
LIMIT 1;

-- name: CreateOrder :one
INSERT INTO "order" (id, price, user_id, payment_id)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, payment_id, price, status, created_at, updated_at;

-- name: UpdateOrder :one
UPDATE "order"
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING id, user_id, payment_id, price, status, created_at, updated_at;
