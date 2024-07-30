package entities

import (
	"database/sql"
)

type VMPostMetrics struct {
	UserID                   sql.NullInt64   `json:"user_id"`
	Name                     sql.NullString  `json:"name"`
	ProfileID                sql.NullInt64   `json:"profile_id"`
	PostID                   sql.NullInt64   `json:"post_id"`
	CreatedAt                sql.NullTime    `json:"created_at"`
	MediaID                  sql.NullInt32   `json:"media_id"`
	HashtagBl                sql.NullString  `json:"hashtag_bl"`
	Likes                    sql.NullInt32   `json:"likes"`
	Comments                 sql.NullInt32   `json:"comments"`
	Shares                   sql.NullInt32   `json:"shares"`
	Views                    sql.NullInt32   `json:"views"`
	Follows                  sql.NullInt32   `json:"follows"`
	EngajamentoMedioPorVideo sql.NullFloat64 `json:"engajamento_medio_por_video"`
}

type PostMetrics struct {
	PostID          sql.NullInt64
	SocialNetworkID sql.NullInt32
	URL             sql.NullString
	ProfileID       sql.NullInt64
	LastUpdated     sql.NullTime
	Likes           sql.NullInt64
	Comments        sql.NullInt64
	Plays           sql.NullInt64
	Shares          sql.NullInt64
}
