package entities

import (
	"database/sql"
)

type ColabMission struct {
	ID                  sql.NullInt32   `json:"id"`
	Title               sql.NullString  `json:"title"`
	Description         sql.NullString  `json:"description"`
	FKIDColabSubmission sql.NullInt32   `json:"fk_id_colab_submission"`
	FKIDBrandDnb        sql.NullInt32   `json:"fk_id_brand_dnb"`
	FKIDColabReward     sql.NullInt32   `json:"fk_id_colab_reward"`
	CreatedAt           sql.NullTime    `json:"created_at"`
	ExpiredAt           sql.NullTime    `json:"expired_at"`
	Status              sql.NullBool    `json:"status"`
	PaymentDays         sql.NullInt32   `json:"payment_days"`
	DeletedAt           sql.NullTime    `json:"deleted_at"`
	ShopifyPriceRuleID  sql.NullString  `json:"shopify_price_rule_id"`
	PriceRuleSetup      sql.NullString  `json:"price_rule_setup"`
	MinimumSubtotal     sql.NullFloat64 `json:"minimum_subtotal"`
	Banner              sql.NullString  `json:"banner"`
	ThumbBanner         sql.NullString  `json:"thumb_banner"`
	RewardDescription   sql.NullString  `json:"reward_description"`
	CumulativeDiscount  sql.NullBool    `json:"cumulative_discount"`
	ChallengeStatusID   sql.NullInt32   `json:"challenge_status_id"`
	MaxValuePerOrder    sql.NullFloat64 `json:"max_value_per_order"`
	UTMSource           sql.NullString  `json:"utm_source"`
}
