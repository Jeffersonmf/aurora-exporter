create schema if not exists social_metrics;

CREATE TABLE social_metrics.v_m_sn_post_hashtags (
  user_id INT,
  name VARCHAR(200),
  profile_id BIGINT,
  post_id BIGINT,
  created_at TIMESTAMP WITHOUT TIME ZONE,
  media_id INT,
  hashtag_bl VARCHAR(600)
);
