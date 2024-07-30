create schema if not exists rds_mysql;

CREATE TABLE rds_mysql.colab_user (
    id SERIAL PRIMARY KEY,
    name varchar(50),
    email varchar(50),
    password varchar(255),
    isactive boolean,
    fk_id_colab_user_invite int,
    termsuse boolean, -- Changed from tinyint(1) to boolean
    emailconfirmationtoken varchar(255),
    token text, -- Changed from longtext to text
    created_at timestamp,
    updated_at timestamp,
    username varchar(30),
    avatar text,
    thumb_avatar text,
    token_push varchar(50),
    creator_size_id int,
    last_social_report_sync timestamp,
    activity text,
    first_access boolean
);