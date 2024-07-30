package entities

import (
	"database/sql"
)

type TotalCouponSales struct {
	BrandID            sql.NullInt32   `json:"brand_id"`
	BrandName          sql.NullString  `json:"brand_name"`
	OrderID            sql.NullString  `json:"order_id"`
	BlCoupon           sql.NullString  `json:"bl_coupon"`
	CouponCreationDate sql.NullTime    `json:"coupon_creation_date"`
	CreatedAt          sql.NullTime    `json:"created_at"`
	UpdatedAt          sql.NullTime    `json:"updated_at"`
	Total              sql.NullFloat64 `json:"total"`
	Discount           sql.NullFloat64 `json:"discount"`
	Status             sql.NullString  `json:"status"`
	FStatus            sql.NullString  `json:"f_status"`
	MissionID          sql.NullInt32   `json:"mission_id"`
	MissionName        sql.NullString  `json:"mission_name"`
	TypeReward         sql.NullInt32   `json:"type_reward"`
	IsApplyTakeRate    sql.NullBool    `json:"is_apply_take_rate"`
	PaymentDays        sql.NullInt32   `json:"payment_days"`
	RewardValue        sql.NullFloat64 `json:"reward_value"`
	UserID             sql.NullString  `json:"user_id"`
	UserName           sql.NullString  `json:"user_name"`
	UserEmail          sql.NullString  `json:"user_email"`
}
