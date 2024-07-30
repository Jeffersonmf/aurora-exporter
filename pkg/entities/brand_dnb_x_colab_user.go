package entities

import "database/sql"

type BrandColabUser struct {
	ID                         sql.NullInt32
	ColabUserID                sql.NullInt32
	BrandDNBID                 sql.NullInt32
	BrandDNBInfluencerInviteID sql.NullInt32
	ApplyTakeRateFlag          sql.NullBool
}
