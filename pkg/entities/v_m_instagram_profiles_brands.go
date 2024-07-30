package entities

import "database/sql"

type VMInstagramProfileBrand struct {
	BrandID                      sql.NullInt32
	InstagramID                  sql.NullString
	InstagramUsername            sql.NullString
	InstagramFullname            sql.NullString
	InstagramIGTVVideoCount      sql.NullInt32
	InstagramPostsCount          sql.NullInt32
	InstagramFollowersCount      sql.NullInt32
	InstagramFollowsCount        sql.NullInt32
	InstagramDataIngestion       sql.NullString
	InstagramBiography           sql.NullString
	InstagramPrivate             sql.NullBool
	InstagramVerified            sql.NullBool
	InstagramProfileURL          sql.NullString
	InstagramProfilePictureURL   sql.NullString
	InstagramProfilePictureURLHD sql.NullString
}
