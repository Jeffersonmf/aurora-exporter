create schema if not exists brandlovrs;

CREATE TABLE brandlovrs.v_m_total_coupon_sales (
    brand_id INT,
    brand_name VARCHAR(255),
    order_id VARCHAR(255),
    bl_coupon VARCHAR(255),
    coupon_creation_date DATE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    total NUMERIC(10, 4) NULL,
    discount NUMERIC(10, 3) NULL,
    status VARCHAR(50),
    f_status VARCHAR(50),
    mission_id INT,
    mission_name VARCHAR(255),
    type_reward INT,
    is_apply_take_rate BOOLEAN,
    payment_days INT,
    reward_value NUMERIC(10, 2) NULL,
    user_id VARCHAR(255),
    user_name VARCHAR(255),
    user_email VARCHAR(255)
);

