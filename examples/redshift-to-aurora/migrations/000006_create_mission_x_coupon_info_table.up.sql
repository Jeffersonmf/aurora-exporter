create schema if not exists rds_mysql;

create table rds_mysql.mission_x_coupon_info
    (
    id int,
    expires_at timestamp,
    total_quantity int,
    brand_prefix varchar(5),
    value decimal(10, 2),
    fk_id_coupon_type_discount int,
    fk_id_colab_mission int,
    max_use_per_creator int
    );