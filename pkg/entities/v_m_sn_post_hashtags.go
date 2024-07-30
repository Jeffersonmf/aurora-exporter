package entities

import (
	"database/sql"
)

type VMsnPostHashtags struct {
	UserID    sql.NullInt64  `json:"user_id"`
	Name      sql.NullString `json:"name"`
	ProfileID sql.NullInt64  `json:"profile_id"`
	PostID    sql.NullInt64  `json:"post_id"`
	CreatedAt sql.NullTime   `json:"created_at"`
	MediaID   sql.NullInt32  `json:"media_id"`
	HashtagBl sql.NullString `json:"hashtag_bl"`
}
