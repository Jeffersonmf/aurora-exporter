create schema if not exists rds_mysql;

CREATE TABLE rds_mysql.brand_dnb_x_influencer_invite (
    id INT,
    fk_id_brand_dnb INT,
    fk_id_colab_user INT,
    acceptedAt TIMESTAMP,
    createdAt TIMESTAMP,
    rejectedAt TIMESTAMP,
    flag_who_invited BOOLEAN,
    flag_apply_take_rate BOOLEAN
);