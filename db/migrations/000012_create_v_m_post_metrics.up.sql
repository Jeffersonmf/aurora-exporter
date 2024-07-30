create schema if not exists social_metrics;

CREATE TABLE social_metrics.v_m_post_metrics (
    post_id BIGSERIAL PRIMARY KEY,
    social_network_id INT,
    url TEXT,
    profile_id BIGINT,
    last_updated TIMESTAMP,
    likes INT,
    comments INT,
    plays INT,
    shares INT
);
