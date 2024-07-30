create schema if not exists brandlovrs;

CREATE TABLE brandlovrs.v_brand_total_balance (
    brand_id INT,
    name VARCHAR(255),
    wallet_credits NUMERIC(18,2),
    provisioned NUMERIC(18,3),
    reserved_debit NUMERIC(18,2),
    reserved NUMERIC(18,2),
    eng_rewards_paid_old NUMERIC(18,2),
    eng_rewards_paid_current NUMERIC(18,2),
    total_random_payments NUMERIC(18,2),
    adjust_debits NUMERIC(18,2),
    available NUMERIC(18,2)
);