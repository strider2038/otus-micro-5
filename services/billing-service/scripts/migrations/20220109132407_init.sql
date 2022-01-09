-- +goose Up
-- +goose StatementBegin
CREATE TABLE "account"
(
    id         uuid primary key,
    amount     float     not null default 0,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

CREATE TABLE "payment"
(
    id         uuid primary key,
    account_id uuid      not null references account (id),
    amount     float     not null,
    created_at timestamp not null default now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "payment";
DROP TABLE "account";
-- +goose StatementEnd
