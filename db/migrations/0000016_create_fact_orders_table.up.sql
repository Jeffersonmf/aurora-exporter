create schema if not exists ecommerce;

CREATE TABLE ecommerce.fact_orders (
    pk_id          INT,
    connector_tag  TEXT,
    workspace_id   UUID,
    source_id      UUID,
    connection_id  UUID,
    order_id       TEXT,
    customer_id    TEXT,
    total          NUMERIC,
    shipping       NUMERIC,
    discount       NUMERIC,
    cancel_reason  TEXT,
    currency       VARCHAR(25),
    internal_id    INT,
    source         TEXT,
    details        TEXT,
    status         TEXT,
    active         BOOLEAN,
    processed_at   TIMESTAMP,
    created_at     TIMESTAMP NULL,
    updated_at     TIMESTAMP NULL,
    deleted_at     TIMESTAMP,
    cancelled_at   TIMESTAMP
);