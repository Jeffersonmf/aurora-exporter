-- Demo data for the local "source" database of docker-compose.yml, which
-- stands in for Redshift (same wire protocol and driver). Not needed
-- against a real cluster: point the pipelines at your own tables.
SET client_min_messages = warning;

CREATE SCHEMA IF NOT EXISTS sales;

DROP TABLE IF EXISTS sales.orders;
DROP TABLE IF EXISTS sales.customers;

CREATE TABLE sales.customers (
    id          bigint PRIMARY KEY,
    name        text NOT NULL,
    email       text,
    country     char(2),
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL
);

CREATE TABLE sales.orders (
    id           bigint PRIMARY KEY,
    customer_id  bigint NOT NULL REFERENCES sales.customers (id),
    status       text NOT NULL,
    total        numeric(14, 2) NOT NULL,
    currency     char(3) NOT NULL,
    created_at   timestamptz NOT NULL,
    updated_at   timestamptz NOT NULL
);

INSERT INTO sales.customers
SELECT i,
       'Customer ' || i,
       CASE WHEN i % 10 = 0 THEN NULL ELSE 'customer' || i || '@example.com' END,
       (ARRAY['BR', 'US', 'PT', 'DE'])[1 + i % 4],
       now() - (i || ' days')::interval,
       now() - (i % 7 || ' days')::interval
FROM generate_series(1, 1000) AS i;

INSERT INTO sales.orders
SELECT i,
       1 + i % 1000,
       (ARRAY['paid', 'shipped', 'delivered', 'refunded'])[1 + i % 4],
       round((10 + random() * 490)::numeric, 2),
       'BRL',
       now() - (i % 90 || ' days')::interval,
       now() - (i % 5 || ' days')::interval
FROM generate_series(1, 20000) AS i;
