create schema if not exists brandlovrs;

CREATE TABLE IF NOT EXISTS brandlovrs.calendar
(
    date TIMESTAMP WITHOUT TIME ZONE,
    year INTEGER,
    month INTEGER,
    day INTEGER,
    dow INTEGER,
    week INTEGER,
    quarter INTEGER
);