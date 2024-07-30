package entities

import (
	"database/sql"
)

type BrandDnbInfluencerInvite struct {
	ID                sql.NullInt64 `json:"id"`
	FkIDBrandDnb      sql.NullInt64 `json:"fk_id_brand_dnb"`
	FkIDColabUser     sql.NullInt64 `json:"fk_id_colab_user"`
	AcceptedAt        sql.NullTime  `json:"acceptedAt"`
	CreatedAt         sql.NullTime  `json:"createdAt"`
	RejectedAt        sql.NullTime  `json:"rejectedAt"`
	FlagWhoInvited    sql.NullBool  `json:"flag_who_invited"`
	FlagApplyTakeRate sql.NullBool  `json:"flag_apply_take_rate"`
}
