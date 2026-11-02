create schema if not exists rds_mysql;

CREATE TABLE rds_mysql.colab_coupon
    (
    id int,
    created_at timestamp,
    expires_at timestamp,
    influencer_coupon_prefix varchar(35),
    quantity int,
    fk_id_colab_user int,
    fk_id_mission_x_coupon_info int,
    ecommerce_coupon_id varchar(60),
    amount_of_use int,
    last_use timestamp
    );