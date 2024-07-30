package entities

import (
	"database/sql"
)

type BrandWalletHistory struct {
	BrandID         sql.NullInt64   `json:"brand_id"`
	BrandName       sql.NullString  `json:"brand_name"`
	MissionID       sql.NullInt64   `json:"mission_id"`
	MissionName     sql.NullString  `json:"mission_name"`
	ChallengeType   sql.NullString  `json:"challenge_type"`
	Value           sql.NullFloat64 `json:"value"`
	OrderID         sql.NullString  `json:"order_id"`
	TransactionDate sql.NullTime    `json:"transaction_date"`
	SourceBalance   sql.NullString  `json:"source_balance"`
	TargetBalance   sql.NullString  `json:"target_balance"`
	TransactionType sql.NullString  `json:"transaction_type"`
	Details         sql.NullString  `json:"details"`
}
