create schema if not exists rds_mysql;

create table rds_mysql.colab_levels_x_missions_x_reward
    (
    fk_id_colab_missions int,
    fk_id_colab_levels int,
    fk_id_colab_type_reward int,
    gross_value decimal(10, 2),
    value decimal(10, 2)
    );