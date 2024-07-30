package entities

import "database/sql"

type VBrandTotalBalance struct {
	BrandID               sql.NullInt32
	Name                  sql.NullString
	WalletCredits         sql.NullString
	Provisioned           sql.NullString
	ReservedDebit         sql.NullString
	Reserved              sql.NullString
	EngRewardsPaidOld     sql.NullString
	EngRewardsPaidCurrent sql.NullString
	TotalRandomPayments   sql.NullString
	AdjustDebits          sql.NullString
	Available             sql.NullString
}
