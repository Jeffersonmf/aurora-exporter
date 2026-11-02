-- Target schema for the quickstart pipelines (golang-migrate format).
-- Column names match the source query aliases: the exporter inserts by
-- column name, so the order of columns here does not matter.
CREATE SCHEMA IF NOT EXISTS analytics;

CREATE TABLE analytics.customers (
    id          bigint       PRIMARY KEY,
    name        text         NOT NULL,
    email       text,
    country     char(2),
    created_at  timestamptz  NOT NULL,
    updated_at  timestamptz  NOT NULL
);

CREATE TABLE analytics.orders (
    id           bigint         PRIMARY KEY,
    customer_id  bigint         NOT NULL,
    status       text           NOT NULL,
    total        numeric(14, 2) NOT NULL,
    currency     char(3)        NOT NULL,
    created_at   timestamptz    NOT NULL,
    updated_at   timestamptz    NOT NULL
);

CREATE INDEX orders_customer_id_idx ON analytics.orders (customer_id);

CREATE TABLE analytics.daily_sales (
    day      date           PRIMARY KEY,
    orders   bigint         NOT NULL,
    revenue  numeric(16, 2) NOT NULL
);
