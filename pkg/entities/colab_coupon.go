package entities

import "database/sql"

type ColabCoupon struct {
	ID                     sql.NullInt32
	CreatedAt              sql.NullTime
	ExpiresAt              sql.NullTime
	InfluencerCouponPrefix sql.NullString
	Quantity               sql.NullInt32
	FkIDColabUser          sql.NullInt32
	FkIDMissionXCouponInfo sql.NullInt32
	EcommerceCouponID      sql.NullString
	AmountOfUse            sql.NullInt32
	LastUse                sql.NullTime
}
