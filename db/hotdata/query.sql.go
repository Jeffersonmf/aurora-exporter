// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package hotdata

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const deleteBrandDnbInHotdata = `-- name: DeleteBrandDnbInHotdata :exec
DELETE FROM rds_mysql.brand_dnb
`

func (q *Queries) DeleteBrandDnbInHotdata(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteBrandDnbInHotdata)
	return err
}

const deleteBrandDnbXColabUser = `-- name: DeleteBrandDnbXColabUser :exec
DELETE FROM rds_mysql.brand_dnb_x_colab_user
`

func (q *Queries) DeleteBrandDnbXColabUser(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteBrandDnbXColabUser)
	return err
}

const deleteBrandDnbXInfluencerInvite = `-- name: DeleteBrandDnbXInfluencerInvite :exec
DELETE FROM rds_mysql.brand_dnb_x_influencer_invite
`

func (q *Queries) DeleteBrandDnbXInfluencerInvite(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteBrandDnbXInfluencerInvite)
	return err
}

const deleteBrandWalletHistory = `-- name: DeleteBrandWalletHistory :exec
DELETE FROM brandlovrs.v_m_brand_wallet_history
`

func (q *Queries) DeleteBrandWalletHistory(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteBrandWalletHistory)
	return err
}

const deleteCalendar = `-- name: DeleteCalendar :exec
DELETE FROM brandlovrs.calendar
`

func (q *Queries) DeleteCalendar(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteCalendar)
	return err
}

const deleteColabCoupon = `-- name: DeleteColabCoupon :exec
DELETE FROM rds_mysql.colab_coupon
`

func (q *Queries) DeleteColabCoupon(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteColabCoupon)
	return err
}

const deleteColabLevelsXMissionsXReward = `-- name: DeleteColabLevelsXMissionsXReward :exec
DELETE FROM rds_mysql.colab_levels_x_missions_x_reward
`

func (q *Queries) DeleteColabLevelsXMissionsXReward(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteColabLevelsXMissionsXReward)
	return err
}

const deleteColabMissionsToHotdata = `-- name: DeleteColabMissionsToHotdata :exec
DELETE FROM rds_mysql.colab_missions
`

func (q *Queries) DeleteColabMissionsToHotdata(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteColabMissionsToHotdata)
	return err
}

const deleteColabMissionsXColabUserToHotdata = `-- name: DeleteColabMissionsXColabUserToHotdata :exec
DELETE FROM rds_mysql.colab_missions_x_colab_user
`

func (q *Queries) DeleteColabMissionsXColabUserToHotdata(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteColabMissionsXColabUserToHotdata)
	return err
}

const deleteColabUserData = `-- name: DeleteColabUserData :exec
DELETE FROM rds_mysql.colab_user_data
`

func (q *Queries) DeleteColabUserData(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteColabUserData)
	return err
}

const deleteColabUserInHotdata = `-- name: DeleteColabUserInHotdata :exec
DELETE FROM rds_mysql.colab_user
`

func (q *Queries) DeleteColabUserInHotdata(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteColabUserInHotdata)
	return err
}

const deleteFactOrders = `-- name: DeleteFactOrders :exec
DELETE FROM ecommerce.fact_orders
`

func (q *Queries) DeleteFactOrders(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteFactOrders)
	return err
}

const deleteInstagramProfilesUsersToHotdata = `-- name: DeleteInstagramProfilesUsersToHotdata :exec
DELETE FROM  brandlovrs.v_m_instagram_profiles_users
`

func (q *Queries) DeleteInstagramProfilesUsersToHotdata(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteInstagramProfilesUsersToHotdata)
	return err
}

const deleteMissionXCouponInfo = `-- name: DeleteMissionXCouponInfo :exec
DELETE FROM rds_mysql.mission_x_coupon_info
`

func (q *Queries) DeleteMissionXCouponInfo(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteMissionXCouponInfo)
	return err
}

const deletePostMetricsToHotdata = `-- name: DeletePostMetricsToHotdata :exec
DELETE FROM social_metrics.v_m_post_metrics
`

func (q *Queries) DeletePostMetricsToHotdata(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deletePostMetricsToHotdata)
	return err
}

const deleteSNPostHashtagsToHotdata = `-- name: DeleteSNPostHashtagsToHotdata :exec
DELETE FROM social_metrics.v_m_sn_post_hashtags
`

func (q *Queries) DeleteSNPostHashtagsToHotdata(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteSNPostHashtagsToHotdata)
	return err
}

const deleteTotalCouponSales = `-- name: DeleteTotalCouponSales :exec
DELETE FROM brandlovrs.v_m_total_coupon_sales
`

func (q *Queries) DeleteTotalCouponSales(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteTotalCouponSales)
	return err
}

const deleteVBrandTotalBalance = `-- name: DeleteVBrandTotalBalance :exec
DELETE FROM brandlovrs.v_brand_total_balance
`

func (q *Queries) DeleteVBrandTotalBalance(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteVBrandTotalBalance)
	return err
}

const deleteVMBrands = `-- name: DeleteVMBrands :exec
DELETE FROM brandlovrs.v_m_brands
`

func (q *Queries) DeleteVMBrands(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteVMBrands)
	return err
}

const deleteVMInstagramProfilesBrands = `-- name: DeleteVMInstagramProfilesBrands :exec
DELETE FROM brandlovrs.v_m_instagram_profiles_brands
`

func (q *Queries) DeleteVMInstagramProfilesBrands(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteVMInstagramProfilesBrands)
	return err
}

const deleteWorkspace = `-- name: DeleteWorkspace :exec
DELETE FROM core.workspace
`

func (q *Queries) DeleteWorkspace(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteWorkspace)
	return err
}

const insertBrandDnbToHotdata = `-- name: InsertBrandDnbToHotdata :one
INSERT INTO rds_mysql.brand_dnb
(id, name, url, brand_banner_url, fk_id_accept_rule, about, created_at, churn_date, churn_finished_at, isactive, datahub_workspace_id, brand_logo_url, brand_banner_url_thumb, number_of_employees_id, average_monthly_revenue_id, brand_logo_url_thumb, minimum_balance, cnpj, fk_id_brand_status, force_apply_take_rate, always_apply_take_rate, company_name)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
RETURNING id
`

type InsertBrandDnbToHotdataParams struct {
	ID                      int32
	Name                    sql.NullString
	Url                     sql.NullString
	BrandBannerUrl          sql.NullString
	FkIDAcceptRule          sql.NullInt32
	About                   sql.NullString
	CreatedAt               sql.NullTime
	ChurnDate               sql.NullTime
	ChurnFinishedAt         sql.NullTime
	Isactive                sql.NullBool
	DatahubWorkspaceID      sql.NullString
	BrandLogoUrl            sql.NullString
	BrandBannerUrlThumb     sql.NullString
	NumberOfEmployeesID     sql.NullInt32
	AverageMonthlyRevenueID sql.NullInt32
	BrandLogoUrlThumb       sql.NullString
	MinimumBalance          sql.NullString
	Cnpj                    sql.NullString
	FkIDBrandStatus         sql.NullInt32
	ForceApplyTakeRate      sql.NullBool
	AlwaysApplyTakeRate     sql.NullBool
	CompanyName             sql.NullString
}

func (q *Queries) InsertBrandDnbToHotdata(ctx context.Context, arg InsertBrandDnbToHotdataParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, insertBrandDnbToHotdata,
		arg.ID,
		arg.Name,
		arg.Url,
		arg.BrandBannerUrl,
		arg.FkIDAcceptRule,
		arg.About,
		arg.CreatedAt,
		arg.ChurnDate,
		arg.ChurnFinishedAt,
		arg.Isactive,
		arg.DatahubWorkspaceID,
		arg.BrandLogoUrl,
		arg.BrandBannerUrlThumb,
		arg.NumberOfEmployeesID,
		arg.AverageMonthlyRevenueID,
		arg.BrandLogoUrlThumb,
		arg.MinimumBalance,
		arg.Cnpj,
		arg.FkIDBrandStatus,
		arg.ForceApplyTakeRate,
		arg.AlwaysApplyTakeRate,
		arg.CompanyName,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const insertBrandDnbXColabUser = `-- name: InsertBrandDnbXColabUser :exec
INSERT INTO rds_mysql.brand_dnb_x_colab_user (id, fk_id_colab_user, fk_id_brand_dnb, fk_id_brand_dnb_x_influencer_invite, flag_apply_take_rate)
VALUES ($1, $2, $3, $4, $5)
`

type InsertBrandDnbXColabUserParams struct {
	ID                            sql.NullInt32
	FkIDColabUser                 sql.NullInt32
	FkIDBrandDnb                  sql.NullInt32
	FkIDBrandDnbXInfluencerInvite sql.NullInt32
	FlagApplyTakeRate             sql.NullBool
}

func (q *Queries) InsertBrandDnbXColabUser(ctx context.Context, arg InsertBrandDnbXColabUserParams) error {
	_, err := q.db.ExecContext(ctx, insertBrandDnbXColabUser,
		arg.ID,
		arg.FkIDColabUser,
		arg.FkIDBrandDnb,
		arg.FkIDBrandDnbXInfluencerInvite,
		arg.FlagApplyTakeRate,
	)
	return err
}

const insertBrandDnbXInfluencerInvite = `-- name: InsertBrandDnbXInfluencerInvite :exec
INSERT INTO rds_mysql.brand_dnb_x_influencer_invite (id, fk_id_brand_dnb, fk_id_colab_user, acceptedAt, createdAt, rejectedAt, flag_who_invited, flag_apply_take_rate)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

type InsertBrandDnbXInfluencerInviteParams struct {
	ID                sql.NullInt32
	FkIDBrandDnb      sql.NullInt32
	FkIDColabUser     sql.NullInt32
	Acceptedat        sql.NullTime
	Createdat         sql.NullTime
	Rejectedat        sql.NullTime
	FlagWhoInvited    sql.NullBool
	FlagApplyTakeRate sql.NullBool
}

func (q *Queries) InsertBrandDnbXInfluencerInvite(ctx context.Context, arg InsertBrandDnbXInfluencerInviteParams) error {
	_, err := q.db.ExecContext(ctx, insertBrandDnbXInfluencerInvite,
		arg.ID,
		arg.FkIDBrandDnb,
		arg.FkIDColabUser,
		arg.Acceptedat,
		arg.Createdat,
		arg.Rejectedat,
		arg.FlagWhoInvited,
		arg.FlagApplyTakeRate,
	)
	return err
}

const insertBrandWalletHistory = `-- name: InsertBrandWalletHistory :exec
INSERT INTO brandlovrs.v_m_brand_wallet_history (
    brand_id,
    brand_name,
    mission_id,
    mission_name,
    challenge_type,
    value,
    order_id,
    transaction_date,
    source_balance,
    target_balance,
    transaction_type,
    details
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
`

type InsertBrandWalletHistoryParams struct {
	BrandID         sql.NullInt32
	BrandName       sql.NullString
	MissionID       sql.NullInt32
	MissionName     sql.NullString
	ChallengeType   sql.NullString
	Value           sql.NullString
	OrderID         sql.NullString
	TransactionDate sql.NullTime
	SourceBalance   sql.NullString
	TargetBalance   sql.NullString
	TransactionType sql.NullString
	Details         sql.NullString
}

func (q *Queries) InsertBrandWalletHistory(ctx context.Context, arg InsertBrandWalletHistoryParams) error {
	_, err := q.db.ExecContext(ctx, insertBrandWalletHistory,
		arg.BrandID,
		arg.BrandName,
		arg.MissionID,
		arg.MissionName,
		arg.ChallengeType,
		arg.Value,
		arg.OrderID,
		arg.TransactionDate,
		arg.SourceBalance,
		arg.TargetBalance,
		arg.TransactionType,
		arg.Details,
	)
	return err
}

const insertCalendar = `-- name: InsertCalendar :exec
INSERT INTO brandlovrs.calendar (date, year, month, day, dow, week, quarter)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

type InsertCalendarParams struct {
	Date    sql.NullTime
	Year    sql.NullInt32
	Month   sql.NullInt32
	Day     sql.NullInt32
	Dow     sql.NullInt32
	Week    sql.NullInt32
	Quarter sql.NullInt32
}

func (q *Queries) InsertCalendar(ctx context.Context, arg InsertCalendarParams) error {
	_, err := q.db.ExecContext(ctx, insertCalendar,
		arg.Date,
		arg.Year,
		arg.Month,
		arg.Day,
		arg.Dow,
		arg.Week,
		arg.Quarter,
	)
	return err
}

const insertColabCoupon = `-- name: InsertColabCoupon :exec
INSERT INTO rds_mysql.colab_coupon (id, created_at, expires_at, influencer_coupon_prefix, quantity, fk_id_colab_user, fk_id_mission_x_coupon_info, ecommerce_coupon_id, amount_of_use, last_use)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

type InsertColabCouponParams struct {
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

func (q *Queries) InsertColabCoupon(ctx context.Context, arg InsertColabCouponParams) error {
	_, err := q.db.ExecContext(ctx, insertColabCoupon,
		arg.ID,
		arg.CreatedAt,
		arg.ExpiresAt,
		arg.InfluencerCouponPrefix,
		arg.Quantity,
		arg.FkIDColabUser,
		arg.FkIDMissionXCouponInfo,
		arg.EcommerceCouponID,
		arg.AmountOfUse,
		arg.LastUse,
	)
	return err
}

const insertColabLevelsXMissionsXReward = `-- name: InsertColabLevelsXMissionsXReward :exec
INSERT INTO rds_mysql.colab_levels_x_missions_x_reward
(fk_id_colab_missions, fk_id_colab_levels, fk_id_colab_type_reward, gross_value, value)
VALUES ($1,$2,$3,$4,$5)
`

type InsertColabLevelsXMissionsXRewardParams struct {
	FkIDColabMissions   sql.NullInt32
	FkIDColabLevels     sql.NullInt32
	FkIDColabTypeReward sql.NullInt32
	GrossValue          sql.NullString
	Value               sql.NullString
}

func (q *Queries) InsertColabLevelsXMissionsXReward(ctx context.Context, arg InsertColabLevelsXMissionsXRewardParams) error {
	_, err := q.db.ExecContext(ctx, insertColabLevelsXMissionsXReward,
		arg.FkIDColabMissions,
		arg.FkIDColabLevels,
		arg.FkIDColabTypeReward,
		arg.GrossValue,
		arg.Value,
	)
	return err
}

const insertColabMissionsToHotdata = `-- name: InsertColabMissionsToHotdata :one
INSERT INTO rds_mysql.colab_missions
(id, title,	"description", fk_id_colab_submission, fk_id_brand_dnb,	fk_id_colab_reward,	created_at,	expired_at,	"status", payment_days,	deleted_at,	shopify_price_rule_id, price_rule_setup, minimum_subtotal, banner, thumb_banner, reward_description, cumulative_discount, challenge_status_id,	max_value_per_order, utm_source)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
RETURNING id
`

type InsertColabMissionsToHotdataParams struct {
	ID                  int32
	Title               sql.NullString
	Description         sql.NullString
	FkIDColabSubmission sql.NullInt32
	FkIDBrandDnb        sql.NullInt32
	FkIDColabReward     sql.NullInt32
	CreatedAt           sql.NullTime
	ExpiredAt           sql.NullTime
	Status              sql.NullBool
	PaymentDays         sql.NullInt32
	DeletedAt           sql.NullTime
	ShopifyPriceRuleID  sql.NullString
	PriceRuleSetup      sql.NullString
	MinimumSubtotal     sql.NullString
	Banner              sql.NullString
	ThumbBanner         sql.NullString
	RewardDescription   sql.NullString
	CumulativeDiscount  sql.NullBool
	ChallengeStatusID   sql.NullInt32
	MaxValuePerOrder    sql.NullString
	UtmSource           sql.NullString
}

func (q *Queries) InsertColabMissionsToHotdata(ctx context.Context, arg InsertColabMissionsToHotdataParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, insertColabMissionsToHotdata,
		arg.ID,
		arg.Title,
		arg.Description,
		arg.FkIDColabSubmission,
		arg.FkIDBrandDnb,
		arg.FkIDColabReward,
		arg.CreatedAt,
		arg.ExpiredAt,
		arg.Status,
		arg.PaymentDays,
		arg.DeletedAt,
		arg.ShopifyPriceRuleID,
		arg.PriceRuleSetup,
		arg.MinimumSubtotal,
		arg.Banner,
		arg.ThumbBanner,
		arg.RewardDescription,
		arg.CumulativeDiscount,
		arg.ChallengeStatusID,
		arg.MaxValuePerOrder,
		arg.UtmSource,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const insertColabMissionsXColabUserToHotdata = `-- name: InsertColabMissionsXColabUserToHotdata :one
INSERT INTO rds_mysql.colab_missions_x_colab_user
(id, fk_id_colab_missions, fk_id_colab_user, "status", flag_apply_take_rate)
VALUES ($1, $2, $3, $4, $5)
RETURNING id
`

type InsertColabMissionsXColabUserToHotdataParams struct {
	ID                int32
	FkIDColabMissions sql.NullInt32
	FkIDColabUser     sql.NullInt32
	Status            sql.NullBool
	FlagApplyTakeRate sql.NullBool
}

func (q *Queries) InsertColabMissionsXColabUserToHotdata(ctx context.Context, arg InsertColabMissionsXColabUserToHotdataParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, insertColabMissionsXColabUserToHotdata,
		arg.ID,
		arg.FkIDColabMissions,
		arg.FkIDColabUser,
		arg.Status,
		arg.FlagApplyTakeRate,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const insertColabUserData = `-- name: InsertColabUserData :exec
INSERT INTO rds_mysql.colab_user_data (id, value, fk_id_colab_type_label, fk_id_colab_user)
VALUES ($1, $2, $3, $4)
`

type InsertColabUserDataParams struct {
	ID                 sql.NullInt32
	Value              sql.NullString
	FkIDColabTypeLabel sql.NullInt32
	FkIDColabUser      sql.NullInt32
}

func (q *Queries) InsertColabUserData(ctx context.Context, arg InsertColabUserDataParams) error {
	_, err := q.db.ExecContext(ctx, insertColabUserData,
		arg.ID,
		arg.Value,
		arg.FkIDColabTypeLabel,
		arg.FkIDColabUser,
	)
	return err
}

const insertColabUserToHotdata = `-- name: InsertColabUserToHotdata :one
INSERT INTO rds_mysql.colab_user
(id, "name", email, "password", isactive, fk_id_colab_user_invite, termsuse, emailconfirmationtoken, "token", created_at, updated_at, username, avatar, thumb_avatar, token_push, creator_size_id, last_social_report_sync, activity, first_access)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
RETURNING id
`

type InsertColabUserToHotdataParams struct {
	ID                     int32
	Name                   sql.NullString
	Email                  sql.NullString
	Password               sql.NullString
	Isactive               sql.NullBool
	FkIDColabUserInvite    sql.NullInt32
	Termsuse               sql.NullBool
	Emailconfirmationtoken sql.NullString
	Token                  sql.NullString
	CreatedAt              sql.NullTime
	UpdatedAt              sql.NullTime
	Username               sql.NullString
	Avatar                 sql.NullString
	ThumbAvatar            sql.NullString
	TokenPush              sql.NullString
	CreatorSizeID          sql.NullInt32
	LastSocialReportSync   sql.NullTime
	Activity               sql.NullString
	FirstAccess            sql.NullBool
}

func (q *Queries) InsertColabUserToHotdata(ctx context.Context, arg InsertColabUserToHotdataParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, insertColabUserToHotdata,
		arg.ID,
		arg.Name,
		arg.Email,
		arg.Password,
		arg.Isactive,
		arg.FkIDColabUserInvite,
		arg.Termsuse,
		arg.Emailconfirmationtoken,
		arg.Token,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Username,
		arg.Avatar,
		arg.ThumbAvatar,
		arg.TokenPush,
		arg.CreatorSizeID,
		arg.LastSocialReportSync,
		arg.Activity,
		arg.FirstAccess,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const insertFactOrder = `-- name: InsertFactOrder :exec
INSERT INTO ecommerce.fact_orders (
			pk_id, connector_tag, workspace_id, source_id, connection_id,
			order_id, customer_id, total, shipping, discount,
			cancel_reason, currency, internal_id, source, details,
			status, active, processed_at, created_at, updated_at,
			deleted_at, cancelled_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
`

type InsertFactOrderParams struct {
	PkID         sql.NullInt32
	ConnectorTag sql.NullString
	WorkspaceID  uuid.NullUUID
	SourceID     uuid.NullUUID
	ConnectionID uuid.NullUUID
	OrderID      sql.NullString
	CustomerID   sql.NullString
	Total        sql.NullString
	Shipping     sql.NullString
	Discount     sql.NullString
	CancelReason sql.NullString
	Currency     sql.NullString
	InternalID   sql.NullInt32
	Source       sql.NullString
	Details      sql.NullString
	Status       sql.NullString
	Active       sql.NullBool
	ProcessedAt  sql.NullTime
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
	DeletedAt    sql.NullTime
	CancelledAt  sql.NullTime
}

func (q *Queries) InsertFactOrder(ctx context.Context, arg InsertFactOrderParams) error {
	_, err := q.db.ExecContext(ctx, insertFactOrder,
		arg.PkID,
		arg.ConnectorTag,
		arg.WorkspaceID,
		arg.SourceID,
		arg.ConnectionID,
		arg.OrderID,
		arg.CustomerID,
		arg.Total,
		arg.Shipping,
		arg.Discount,
		arg.CancelReason,
		arg.Currency,
		arg.InternalID,
		arg.Source,
		arg.Details,
		arg.Status,
		arg.Active,
		arg.ProcessedAt,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.DeletedAt,
		arg.CancelledAt,
	)
	return err
}

const insertInstagramProfilesUsersToHotdata = `-- name: InsertInstagramProfilesUsersToHotdata :one
INSERT INTO brandlovrs.v_m_instagram_profiles_users
(user_id, instagram_username, instagram_fullname, instagram_id, business_category, instagram_igtvvideocount, instagram_postscount, instagram_followerscount, instagram_followscount, instagram_data_ingestion, instagram_biography, instagram_private, instagram_verified, instagram_profile_url, instagram_profile_picture_url, instagram_profile_picture_url_hd)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING user_id
`

type InsertInstagramProfilesUsersToHotdataParams struct {
	UserID                       sql.NullInt32
	InstagramUsername            sql.NullString
	InstagramFullname            sql.NullString
	InstagramID                  sql.NullString
	BusinessCategory             sql.NullString
	InstagramIgtvvideocount      sql.NullInt32
	InstagramPostscount          sql.NullInt32
	InstagramFollowerscount      sql.NullInt32
	InstagramFollowscount        sql.NullInt32
	InstagramDataIngestion       sql.NullString
	InstagramBiography           sql.NullString
	InstagramPrivate             sql.NullBool
	InstagramVerified            sql.NullBool
	InstagramProfileUrl          sql.NullString
	InstagramProfilePictureUrl   sql.NullString
	InstagramProfilePictureUrlHd sql.NullString
}

func (q *Queries) InsertInstagramProfilesUsersToHotdata(ctx context.Context, arg InsertInstagramProfilesUsersToHotdataParams) (sql.NullInt32, error) {
	row := q.db.QueryRowContext(ctx, insertInstagramProfilesUsersToHotdata,
		arg.UserID,
		arg.InstagramUsername,
		arg.InstagramFullname,
		arg.InstagramID,
		arg.BusinessCategory,
		arg.InstagramIgtvvideocount,
		arg.InstagramPostscount,
		arg.InstagramFollowerscount,
		arg.InstagramFollowscount,
		arg.InstagramDataIngestion,
		arg.InstagramBiography,
		arg.InstagramPrivate,
		arg.InstagramVerified,
		arg.InstagramProfileUrl,
		arg.InstagramProfilePictureUrl,
		arg.InstagramProfilePictureUrlHd,
	)
	var user_id sql.NullInt32
	err := row.Scan(&user_id)
	return user_id, err
}

const insertMissionXCouponInfo = `-- name: InsertMissionXCouponInfo :exec
INSERT INTO rds_mysql.mission_x_coupon_info(
    id, expires_at, total_quantity, brand_prefix, value, fk_id_coupon_type_discount,
    fk_id_colab_mission, max_use_per_creator)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

type InsertMissionXCouponInfoParams struct {
	ID                     sql.NullInt32
	ExpiresAt              sql.NullTime
	TotalQuantity          sql.NullInt32
	BrandPrefix            sql.NullString
	Value                  sql.NullString
	FkIDCouponTypeDiscount sql.NullInt32
	FkIDColabMission       sql.NullInt32
	MaxUsePerCreator       sql.NullInt32
}

func (q *Queries) InsertMissionXCouponInfo(ctx context.Context, arg InsertMissionXCouponInfoParams) error {
	_, err := q.db.ExecContext(ctx, insertMissionXCouponInfo,
		arg.ID,
		arg.ExpiresAt,
		arg.TotalQuantity,
		arg.BrandPrefix,
		arg.Value,
		arg.FkIDCouponTypeDiscount,
		arg.FkIDColabMission,
		arg.MaxUsePerCreator,
	)
	return err
}

const insertPostMetricsToHotdata = `-- name: InsertPostMetricsToHotdata :exec
INSERT INTO social_metrics.v_m_post_metrics (post_id, social_network_id, url, profile_id, last_updated, likes, comments, plays, shares)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type InsertPostMetricsToHotdataParams struct {
	PostID          int64
	SocialNetworkID sql.NullInt32
	Url             sql.NullString
	ProfileID       sql.NullInt64
	LastUpdated     sql.NullTime
	Likes           sql.NullInt32
	Comments        sql.NullInt32
	Plays           sql.NullInt32
	Shares          sql.NullInt32
}

func (q *Queries) InsertPostMetricsToHotdata(ctx context.Context, arg InsertPostMetricsToHotdataParams) error {
	_, err := q.db.ExecContext(ctx, insertPostMetricsToHotdata,
		arg.PostID,
		arg.SocialNetworkID,
		arg.Url,
		arg.ProfileID,
		arg.LastUpdated,
		arg.Likes,
		arg.Comments,
		arg.Plays,
		arg.Shares,
	)
	return err
}

const insertSNPostHashtagsToHotdata = `-- name: InsertSNPostHashtagsToHotdata :one
INSERT INTO social_metrics.v_m_sn_post_hashtags
(user_id, "name", profile_id, post_id, created_at, media_id, hashtag_bl)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING user_id
`

type InsertSNPostHashtagsToHotdataParams struct {
	UserID    sql.NullInt32
	Name      sql.NullString
	ProfileID sql.NullInt64
	PostID    sql.NullInt64
	CreatedAt sql.NullTime
	MediaID   sql.NullInt32
	HashtagBl sql.NullString
}

func (q *Queries) InsertSNPostHashtagsToHotdata(ctx context.Context, arg InsertSNPostHashtagsToHotdataParams) (sql.NullInt32, error) {
	row := q.db.QueryRowContext(ctx, insertSNPostHashtagsToHotdata,
		arg.UserID,
		arg.Name,
		arg.ProfileID,
		arg.PostID,
		arg.CreatedAt,
		arg.MediaID,
		arg.HashtagBl,
	)
	var user_id sql.NullInt32
	err := row.Scan(&user_id)
	return user_id, err
}

const insertTotalCouponSales = `-- name: InsertTotalCouponSales :exec
INSERT INTO brandlovrs.v_m_total_coupon_sales (
    brand_id,
    brand_name,
    order_id,
    bl_coupon,
    coupon_creation_date,
    created_at,
    updated_at,
    total,
    discount,
    status,
    f_status,
    mission_id,
    mission_name,
    type_reward,
    is_apply_take_rate,
    payment_days,
    reward_value,
    user_id,
    user_name,
    user_email
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17,
    $18,
    $19,
    $20
)
`

type InsertTotalCouponSalesParams struct {
	BrandID            sql.NullInt32
	BrandName          sql.NullString
	OrderID            sql.NullString
	BlCoupon           sql.NullString
	CouponCreationDate sql.NullTime
	CreatedAt          sql.NullTime
	UpdatedAt          sql.NullTime
	Total              sql.NullString
	Discount           sql.NullString
	Status             sql.NullString
	FStatus            sql.NullString
	MissionID          sql.NullInt32
	MissionName        sql.NullString
	TypeReward         sql.NullInt32
	IsApplyTakeRate    sql.NullBool
	PaymentDays        sql.NullInt32
	RewardValue        sql.NullString
	UserID             sql.NullString
	UserName           sql.NullString
	UserEmail          sql.NullString
}

func (q *Queries) InsertTotalCouponSales(ctx context.Context, arg InsertTotalCouponSalesParams) error {
	_, err := q.db.ExecContext(ctx, insertTotalCouponSales,
		arg.BrandID,
		arg.BrandName,
		arg.OrderID,
		arg.BlCoupon,
		arg.CouponCreationDate,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Total,
		arg.Discount,
		arg.Status,
		arg.FStatus,
		arg.MissionID,
		arg.MissionName,
		arg.TypeReward,
		arg.IsApplyTakeRate,
		arg.PaymentDays,
		arg.RewardValue,
		arg.UserID,
		arg.UserName,
		arg.UserEmail,
	)
	return err
}

const insertVBrandTotalBalance = `-- name: InsertVBrandTotalBalance :exec
INSERT INTO brandlovrs.v_brand_total_balance (
    brand_id,
    name,
    wallet_credits,
    provisioned,
    reserved_debit,
    reserved,
    eng_rewards_paid_old,
    eng_rewards_paid_current,
    total_random_payments,
    adjust_debits,
    available) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`

type InsertVBrandTotalBalanceParams struct {
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

func (q *Queries) InsertVBrandTotalBalance(ctx context.Context, arg InsertVBrandTotalBalanceParams) error {
	_, err := q.db.ExecContext(ctx, insertVBrandTotalBalance,
		arg.BrandID,
		arg.Name,
		arg.WalletCredits,
		arg.Provisioned,
		arg.ReservedDebit,
		arg.Reserved,
		arg.EngRewardsPaidOld,
		arg.EngRewardsPaidCurrent,
		arg.TotalRandomPayments,
		arg.AdjustDebits,
		arg.Available,
	)
	return err
}

const insertVMBrands = `-- name: InsertVMBrands :exec
INSERT INTO brandlovrs.v_m_brands (
    brand_id,
    brand,
    cnpj,
    company_name,
    interest,
    challenges,
    active_challenges,
    total_ambassadors,
    ambassadors_with_sales,
    total_sales,
    brandlovrs_rate,
    gmv,
    gmv_7,
    gmv_15,
    gmv_30,
    gmv_60,
    gmv_90,
    last_sale,
    provisioned_balance,
    reserved_balance,
    available_balance,
    created_at,
    status,
    brand_activity,
    brand_logo_url
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17,
    $18,
    $19,
    $20,
    $21,
    $22,
    $23,
    $24,
    $25
)
`

type InsertVMBrandsParams struct {
	BrandID              sql.NullInt32
	Brand                sql.NullString
	Cnpj                 sql.NullString
	CompanyName          sql.NullString
	Interest             sql.NullString
	Challenges           sql.NullInt32
	ActiveChallenges     sql.NullInt32
	TotalAmbassadors     sql.NullInt32
	AmbassadorsWithSales sql.NullInt32
	TotalSales           sql.NullString
	BrandlovrsRate       sql.NullString
	Gmv                  sql.NullString
	Gmv7                 sql.NullString
	Gmv15                sql.NullString
	Gmv30                sql.NullString
	Gmv60                sql.NullString
	Gmv90                sql.NullString
	LastSale             sql.NullTime
	ProvisionedBalance   sql.NullString
	ReservedBalance      sql.NullString
	AvailableBalance     sql.NullString
	CreatedAt            sql.NullTime
	Status               sql.NullBool
	BrandActivity        sql.NullString
	BrandLogoUrl         sql.NullString
}

func (q *Queries) InsertVMBrands(ctx context.Context, arg InsertVMBrandsParams) error {
	_, err := q.db.ExecContext(ctx, insertVMBrands,
		arg.BrandID,
		arg.Brand,
		arg.Cnpj,
		arg.CompanyName,
		arg.Interest,
		arg.Challenges,
		arg.ActiveChallenges,
		arg.TotalAmbassadors,
		arg.AmbassadorsWithSales,
		arg.TotalSales,
		arg.BrandlovrsRate,
		arg.Gmv,
		arg.Gmv7,
		arg.Gmv15,
		arg.Gmv30,
		arg.Gmv60,
		arg.Gmv90,
		arg.LastSale,
		arg.ProvisionedBalance,
		arg.ReservedBalance,
		arg.AvailableBalance,
		arg.CreatedAt,
		arg.Status,
		arg.BrandActivity,
		arg.BrandLogoUrl,
	)
	return err
}

const insertVMInstagramProfilesBrands = `-- name: InsertVMInstagramProfilesBrands :exec
INSERT INTO brandlovrs.v_m_instagram_profiles_brands (brand_id, instagram_id, instagram_username, instagram_fullname, instagram_igtvvideocount, instagram_postscount, instagram_followerscount, instagram_followscount, instagram_data_ingestion, instagram_biography, instagram_private, instagram_verified, instagram_profile_url, instagram_profile_picture_url, instagram_profile_picture_url_hd)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`

type InsertVMInstagramProfilesBrandsParams struct {
	BrandID                      sql.NullInt32
	InstagramID                  sql.NullString
	InstagramUsername            sql.NullString
	InstagramFullname            sql.NullString
	InstagramIgtvvideocount      sql.NullInt32
	InstagramPostscount          sql.NullInt32
	InstagramFollowerscount      sql.NullInt32
	InstagramFollowscount        sql.NullInt32
	InstagramDataIngestion       sql.NullString
	InstagramBiography           sql.NullString
	InstagramPrivate             sql.NullBool
	InstagramVerified            sql.NullBool
	InstagramProfileUrl          sql.NullString
	InstagramProfilePictureUrl   sql.NullString
	InstagramProfilePictureUrlHd sql.NullString
}

func (q *Queries) InsertVMInstagramProfilesBrands(ctx context.Context, arg InsertVMInstagramProfilesBrandsParams) error {
	_, err := q.db.ExecContext(ctx, insertVMInstagramProfilesBrands,
		arg.BrandID,
		arg.InstagramID,
		arg.InstagramUsername,
		arg.InstagramFullname,
		arg.InstagramIgtvvideocount,
		arg.InstagramPostscount,
		arg.InstagramFollowerscount,
		arg.InstagramFollowscount,
		arg.InstagramDataIngestion,
		arg.InstagramBiography,
		arg.InstagramPrivate,
		arg.InstagramVerified,
		arg.InstagramProfileUrl,
		arg.InstagramProfilePictureUrl,
		arg.InstagramProfilePictureUrlHd,
	)
	return err
}

const insertWorkspace = `-- name: InsertWorkspace :exec
INSERT INTO core.workspace (id, customer_id, email, name, reference_id, integrator_id, active, created_at, updated_at, type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

type InsertWorkspaceParams struct {
	ID           uuid.NullUUID
	CustomerID   uuid.NullUUID
	Email        sql.NullString
	Name         sql.NullString
	ReferenceID  sql.NullString
	IntegratorID sql.NullInt32
	Active       sql.NullBool
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
	Type         sql.NullString
}

func (q *Queries) InsertWorkspace(ctx context.Context, arg InsertWorkspaceParams) error {
	_, err := q.db.ExecContext(ctx, insertWorkspace,
		arg.ID,
		arg.CustomerID,
		arg.Email,
		arg.Name,
		arg.ReferenceID,
		arg.IntegratorID,
		arg.Active,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Type,
	)
	return err
}
