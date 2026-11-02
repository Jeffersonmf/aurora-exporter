create schema if not exists core;

CREATE TABLE core.workspace (
    id  UUID,
    customer_id UUID,
    email VARCHAR(255),
    name VARCHAR(255),
    reference_id VARCHAR(255),
    integrator_id INT,
    active BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    type VARCHAR(100)
);
