create schema if not exists rds_mysql;

CREATE TABLE rds_mysql.colab_user_data (
    id INT,
    value TEXT,
    fk_id_colab_type_label INT,
    fk_id_colab_user INT
);
