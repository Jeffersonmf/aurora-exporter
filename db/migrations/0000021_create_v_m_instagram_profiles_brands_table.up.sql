create schema if not exists brandlovrs;

CREATE TABLE IF NOT EXISTS brandlovrs.v_m_instagram_profiles_brands
(
    brand_id INTEGER,
    instagram_id TEXT,
    instagram_username TEXT,
    instagram_fullname TEXT,
    instagram_igtvvideocount INTEGER,
    instagram_postscount INTEGER,
    instagram_followerscount INTEGER,
    instagram_followscount INTEGER,
    instagram_data_ingestion TEXT,
    instagram_biography TEXT,
    instagram_private BOOLEAN,
    instagram_verified BOOLEAN,
    instagram_profile_url TEXT,
    instagram_profile_picture_url TEXT,
    instagram_profile_picture_url_hd TEXT
);