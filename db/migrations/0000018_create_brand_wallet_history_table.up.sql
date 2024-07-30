create schema if not exists brandlovrs;

CREATE TABLE brandlovrs.v_m_brand_wallet_history (
    brand_id         INT,
    brand_name       VARCHAR(100),
    mission_id       INT,
    mission_name     VARCHAR(255),
    challenge_type   VARCHAR(50),
    value            NUMERIC(12, 2), -- Assuming currency values
    order_id         VARCHAR(50),    -- Assuming order IDs can have alphanumeric characters
    transaction_date TIMESTAMP,
    source_balance   VARCHAR(50),    -- Assuming balance values can have alphanumeric characters
    target_balance   VARCHAR(50),    -- Assuming balance values can have alphanumeric characters
    transaction_type VARCHAR(50),
    details          TEXT
);