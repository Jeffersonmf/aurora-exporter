create schema if not exists brandlovrs;

CREATE TABLE brandlovrs.v_m_brands (
    brand_id               int,
    brand                  text,
    cnpj                   varchar(72),
    company_name           text,
    interest               text,
    challenges             int,
    active_challenges      int,
    total_ambassadors      int,
    ambassadors_with_sales int,
    total_sales            decimal(16, 2),
    brandlovrs_rate        decimal(16, 2),
    gmv                    decimal(16, 2),
    gmv_7                  decimal(16, 2),
    gmv_15                 decimal(16, 2),
    gmv_30                 decimal(16, 2),
    gmv_60                 decimal(16, 2),
    gmv_90                 decimal(16, 2),
    last_sale              timestamp ,
    provisioned_balance    decimal(16, 2),
    reserved_balance       decimal(16, 2),
    available_balance      decimal(16, 2),
    created_at             timestamp,
    status                 boolean,
    brand_activity         varchar(150),
    brand_logo_url         text
);

