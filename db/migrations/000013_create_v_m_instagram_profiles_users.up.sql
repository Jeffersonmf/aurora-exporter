create schema if not exists brandlovrs;

CREATE TABLE brandlovrs.v_m_instagram_profiles_users (
  user_id INT,
  instagram_id VARCHAR(16383),
  instagram_username VARCHAR(16383),
  instagram_fullname VARCHAR(16383),
  business_category VARCHAR(16383),
  instagram_igtvvideocount INT,
  instagram_postscount INT,
  instagram_followerscount INT,
  instagram_followscount INT,
  instagram_data_ingestion VARCHAR(16383),
  instagram_biography VARCHAR(16383),
  instagram_private BOOLEAN,
  instagram_verified BOOLEAN,
  instagram_profile_url VARCHAR(16383),
  instagram_profile_picture_url VARCHAR(16383),
  instagram_profile_picture_url_hd VARCHAR(16383)
  );
