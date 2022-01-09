CREATE TABLE "user"
(
    id         bigserial primary key,
    email      text not null unique,
    password   text not null,
    first_name text not null,
    last_name  text not null,
    phone      text not null default ''
)
