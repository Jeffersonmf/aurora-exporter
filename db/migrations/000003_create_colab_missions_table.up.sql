create schema if not exists rds_mysql;

CREATE TABLE rds_mysql.colab_missions (
    id SERIAL PRIMARY KEY,
    title varchar(800),
    description text,
    fk_id_colab_submission int,
    fk_id_brand_dnb int,
    fk_id_colab_reward int,
    created_at timestamp,
    expired_at timestamp,
    status boolean,
    payment_days int,
    deleted_at timestamp,
    shopify_price_rule_id varchar(45),
    price_rule_setup text,
    minimum_subtotal decimal(10, 2),
    banner text,
    thumb_banner varchar(700),
    reward_description text,
    cumulative_discount boolean,
    challenge_status_id int,
    max_value_per_order decimal(10, 2),
    utm_source varchar(255)
);


