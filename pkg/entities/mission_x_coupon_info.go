package entities

import "database/sql"

type MissionXCouponInfo struct {
	ID                     int             `json:"id"`
	ExpiresAt              sql.NullTime    `json:"expires_at"`
	TotalQuantity          sql.NullInt32   `json:"total_quantity"`
	BrandPrefix            sql.NullString  `json:"brand_prefix"`
	Value                  sql.NullFloat64 `json:"value"`
	FkIDCouponTypeDiscount sql.NullInt32   `json:"fk_id_coupon_type_discount"`
	FkIDColabMission       sql.NullInt32   `json:"fk_id_colab_mission"`
	MaxUsePerCreator       sql.NullInt32   `json:"max_use_per_creator"`
}
