package entities

import (
	"database/sql"
)

type VMInstagramProfilesUsers struct {
	UserID                       sql.NullInt64  `json:"user_id"`
	InstagramID                  sql.NullString `json:"instagram_id"`
	InstagramUsername            sql.NullString `json:"instagram_username"`
	InstagramFullname            sql.NullString `json:"instagram_fullname"`
	BusinessCategory             sql.NullString `json:"business_category"`
	InstagramIGTVVideoCount      sql.NullInt64  `json:"instagram_igtvvideocount"`
	InstagramPostsCount          sql.NullInt64  `json:"instagram_postscount"`
	InstagramFollowersCount      sql.NullInt64  `json:"instagram_followerscount"`
	InstagramFollowsCount        sql.NullInt64  `json:"instagram_followscount"`
	InstagramDataIngestion       sql.NullString `json:"instagram_data_ingestion"`
	InstagramBiography           sql.NullString `json:"instagram_biography"`
	InstagramPrivate             sql.NullBool   `json:"instagram_private"`
	InstagramVerified            sql.NullBool   `json:"instagram_verified"`
	InstagramProfileURL          sql.NullString `json:"instagram_profile_url"`
	InstagramProfilePictureURL   sql.NullString `json:"instagram_profile_picture_url"`
	InstagramProfilePictureURLHD sql.NullString `json:"instagram_profile_picture_url_hd"`
}
