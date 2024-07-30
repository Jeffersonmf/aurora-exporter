create schema if not exists rds_mysql;

CREATE TABLE rds_mysql.colab_missions_x_colab_user (
    id SERIAL PRIMARY KEY,
    fk_id_colab_missions int,
    fk_id_colab_user int,
    status boolean,
    flag_apply_take_rate boolean
);
