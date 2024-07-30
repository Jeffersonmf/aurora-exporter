create schema if not exists datahub;

CREATE TABLE datahub.execution_history_logs (
    id UUID,
    table_name varchar(255),
    executed_at TIMESTAMP,
    total_records bigint
);