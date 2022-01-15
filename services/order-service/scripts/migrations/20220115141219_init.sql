-- +goose Up
-- +goose StatementBegin
CREATE TABLE "order"
(
    id         uuid primary key,
    user_id    uuid      not null,
    payment_id uuid      not null unique,
    price      float     not null,
    status     text      not null default 'pending',
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "order";
-- +goose StatementEnd
