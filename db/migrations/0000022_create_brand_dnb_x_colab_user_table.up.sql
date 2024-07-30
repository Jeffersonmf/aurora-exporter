create schema if not exists rds_mysql;

CREATE TABLE rds_mysql.brand_dnb_x_colab_user
(
    id INTEGER,
    fk_id_colab_user INTEGER,
    fk_id_brand_dnb INTEGER,
    fk_id_brand_dnb_x_influencer_invite INTEGER,
    flag_apply_take_rate BOOLEAN
);