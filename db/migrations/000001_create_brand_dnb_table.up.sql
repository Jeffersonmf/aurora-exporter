create schema if not exists rds_mysql;

CREATE TABLE rds_mysql.brand_dnb (
    id SERIAL PRIMARY KEY,
    name varchar(100),
    url varchar(400),
    brand_banner_url varchar(100),
    fk_id_accept_rule int,
    about text,
    created_at timestamp,
    churn_date date,
    churn_finished_at timestamp,
    isactive boolean,
    datahub_workspace_id varchar(100),
    brand_logo_url varchar(255),
    brand_banner_url_thumb varchar(100),
    number_of_employees_id int,
    average_monthly_revenue_id int,
    brand_logo_url_thumb varchar(700),
    minimum_balance numeric(10, 2),
    cnpj varchar(18),
    fk_id_brand_status int,
    force_apply_take_rate boolean,
    always_apply_take_rate boolean,
    company_name varchar(255)
);


