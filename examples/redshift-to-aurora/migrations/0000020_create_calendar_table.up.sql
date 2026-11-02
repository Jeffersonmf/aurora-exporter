create schema if not exists analytics;

CREATE TABLE IF NOT EXISTS analytics.calendar
(
    date TIMESTAMP WITHOUT TIME ZONE,
    year INTEGER,
    month INTEGER,
    day INTEGER,
    dow INTEGER,
    week INTEGER,
    quarter INTEGER
);