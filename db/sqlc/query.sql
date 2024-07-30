-- name: InsertBrandDnbToHotdata :one
INSERT INTO rds_mysql.brand_dnb
(id, name, url, brand_banner_url, fk_id_accept_rule, about, created_at, churn_date, churn_finished_at, isactive, datahub_workspace_id, brand_logo_url, brand_banner_url_thumb, number_of_employees_id, average_monthly_revenue_id, brand_logo_url_thumb, minimum_balance, cnpj, fk_id_brand_status, force_apply_take_rate, always_apply_take_rate, company_name)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
RETURNING id;

-- name: DeleteBrandDnbInHotdata :exec
DELETE FROM rds_mysql.brand_dnb;

-- name: InsertColabUserToHotdata :one
INSERT INTO rds_mysql.colab_user
(id, "name", email, "password", isactive, fk_id_colab_user_invite, termsuse, emailconfirmationtoken, "token", created_at, updated_at, username, avatar, thumb_avatar, token_push, creator_size_id, last_social_report_sync, activity, first_access)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
RETURNING id;

-- name: DeleteColabUserInHotdata :exec
DELETE FROM rds_mysql.colab_user;

-- name: InsertColabMissionsToHotdata :one
INSERT INTO rds_mysql.colab_missions
(id, title,	"description", fk_id_colab_submission, fk_id_brand_dnb,	fk_id_colab_reward,	created_at,	expired_at,	"status", payment_days,	deleted_at,	shopify_price_rule_id, price_rule_setup, minimum_subtotal, banner, thumb_banner, reward_description, cumulative_discount, challenge_status_id,	max_value_per_order, utm_source)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
RETURNING id;

-- name: DeleteColabMissionsToHotdata :exec
DELETE FROM rds_mysql.colab_missions;

-- name: InsertColabMissionsXColabUserToHotdata :one
INSERT INTO rds_mysql.colab_missions_x_colab_user
(id, fk_id_colab_missions, fk_id_colab_user, "status", flag_apply_take_rate)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: DeleteColabMissionsXColabUserToHotdata :exec
DELETE FROM rds_mysql.colab_missions_x_colab_user;

-- name: InsertSNPostHashtagsToHotdata :one
INSERT INTO social_metrics.v_m_sn_post_hashtags
(user_id, "name", profile_id, post_id, created_at, media_id, hashtag_bl)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING user_id;

-- name: DeleteSNPostHashtagsToHotdata :exec
DELETE FROM social_metrics.v_m_sn_post_hashtags;

-- name: InsertPostMetricsToHotdata :exec
INSERT INTO social_metrics.v_m_post_metrics (post_id, social_network_id, url, profile_id, last_updated, likes, comments, plays, shares)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: DeletePostMetricsToHotdata :exec
DELETE FROM social_metrics.v_m_post_metrics;

-- name: InsertInstagramProfilesUsersToHotdata :one
INSERT INTO brandlovrs.v_m_instagram_profiles_users
(user_id, instagram_username, instagram_fullname, instagram_id, business_category, instagram_igtvvideocount, instagram_postscount, instagram_followerscount, instagram_followscount, instagram_data_ingestion, instagram_biography, instagram_private, instagram_verified, instagram_profile_url, instagram_profile_picture_url, instagram_profile_picture_url_hd)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING user_id;

-- name: DeleteInstagramProfilesUsersToHotdata :exec
DELETE FROM  brandlovrs.v_m_instagram_profiles_users;

-- name: InsertColabLevelsXMissionsXReward :exec
INSERT INTO rds_mysql.colab_levels_x_missions_x_reward
(fk_id_colab_missions, fk_id_colab_levels, fk_id_colab_type_reward, gross_value, value)
VALUES ($1,$2,$3,$4,$5);

-- name: DeleteColabLevelsXMissionsXReward :exec
DELETE FROM rds_mysql.colab_levels_x_missions_x_reward;

-- name: InsertColabCoupon :exec
INSERT INTO rds_mysql.colab_coupon (id, created_at, expires_at, influencer_coupon_prefix, quantity, fk_id_colab_user, fk_id_mission_x_coupon_info, ecommerce_coupon_id, amount_of_use, last_use)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: DeleteColabCoupon :exec
DELETE FROM rds_mysql.colab_coupon;

-- name: InsertMissionXCouponInfo :exec
INSERT INTO rds_mysql.mission_x_coupon_info(
    id, expires_at, total_quantity, brand_prefix, value, fk_id_coupon_type_discount,
    fk_id_colab_mission, max_use_per_creator)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: DeleteMissionXCouponInfo :exec
DELETE FROM rds_mysql.mission_x_coupon_info;

-- name: InsertVMBrands :exec
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
);

-- name: DeleteVMBrands :exec
DELETE FROM brandlovrs.v_m_brands;

-- name: InsertTotalCouponSales :exec
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
);

-- name: DeleteTotalCouponSales :exec
DELETE FROM brandlovrs.v_m_total_coupon_sales;

-- name: InsertVBrandTotalBalance :exec
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
    available) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: DeleteVBrandTotalBalance :exec
DELETE FROM brandlovrs.v_brand_total_balance;

-- name: InsertFactOrder :exec
INSERT INTO ecommerce.fact_orders (
			pk_id, connector_tag, workspace_id, source_id, connection_id,
			order_id, customer_id, total, shipping, discount,
			cancel_reason, currency, internal_id, source, details,
			status, active, processed_at, created_at, updated_at,
			deleted_at, cancelled_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22);

-- name: DeleteFactOrders :exec
DELETE FROM ecommerce.fact_orders; 

-- name: InsertBrandWalletHistory :exec
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
);

-- name: DeleteBrandWalletHistory :exec
DELETE FROM brandlovrs.v_m_brand_wallet_history; 

-- name: InsertBrandDnbXInfluencerInvite :exec
INSERT INTO rds_mysql.brand_dnb_x_influencer_invite (id, fk_id_brand_dnb, fk_id_colab_user, acceptedAt, createdAt, rejectedAt, flag_who_invited, flag_apply_take_rate)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: DeleteBrandDnbXInfluencerInvite :exec
DELETE FROM rds_mysql.brand_dnb_x_influencer_invite; 

-- name: InsertColabUserData :exec
INSERT INTO rds_mysql.colab_user_data (id, value, fk_id_colab_type_label, fk_id_colab_user)
VALUES ($1, $2, $3, $4);

-- name: DeleteColabUserData :exec
DELETE FROM rds_mysql.colab_user_data; 

-- name: InsertWorkspace :exec
INSERT INTO core.workspace (id, customer_id, email, name, reference_id, integrator_id, active, created_at, updated_at, type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: DeleteWorkspace :exec
DELETE FROM core.workspace; 

-- name: InsertCalendar :exec
INSERT INTO brandlovrs.calendar (date, year, month, day, dow, week, quarter)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: DeleteCalendar :exec
DELETE FROM brandlovrs.calendar; 

-- name: InsertVMInstagramProfilesBrands :exec
INSERT INTO brandlovrs.v_m_instagram_profiles_brands (brand_id, instagram_id, instagram_username, instagram_fullname, instagram_igtvvideocount, instagram_postscount, instagram_followerscount, instagram_followscount, instagram_data_ingestion, instagram_biography, instagram_private, instagram_verified, instagram_profile_url, instagram_profile_picture_url, instagram_profile_picture_url_hd)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: DeleteVMInstagramProfilesBrands :exec
DELETE FROM brandlovrs.v_m_instagram_profiles_brands; 

-- name: InsertBrandDnbXColabUser :exec
INSERT INTO rds_mysql.brand_dnb_x_colab_user (id, fk_id_colab_user, fk_id_brand_dnb, fk_id_brand_dnb_x_influencer_invite, flag_apply_take_rate)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteBrandDnbXColabUser :exec
DELETE FROM rds_mysql.brand_dnb_x_colab_user; 