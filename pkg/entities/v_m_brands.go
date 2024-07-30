package entities

import "database/sql"

type VMBrands struct {
	BrandID              sql.NullInt64   `json:"brand_id"`
	Brand                sql.NullString  `json:"brand"`
	CNPJ                 sql.NullString  `json:"cnpj"`
	CompanyName          sql.NullString  `json:"company_name"`
	Interest             sql.NullString  `json:"interest"`
	Challenges           sql.NullInt64   `json:"challenges"`
	ActiveChallenges     sql.NullInt64   `json:"active_challenges"`
	TotalAmbassadors     sql.NullInt64   `json:"total_ambassadors"`
	AmbassadorsWithSales sql.NullInt64   `json:"ambassadors_with_sales"`
	TotalSales           sql.NullFloat64 `json:"total_sales"`
	BrandlovrsRate       sql.NullFloat64 `json:"brandlovrs_rate"`
	GMV                  sql.NullFloat64 `json:"gmv"`
	GMV7                 sql.NullFloat64 `json:"gmv_7"`
	GMV15                sql.NullFloat64 `json:"gmv_15"`
	GMV30                sql.NullFloat64 `json:"gmv_30"`
	GMV60                sql.NullFloat64 `json:"gmv_60"`
	GMV90                sql.NullFloat64 `json:"gmv_90"`
	LastSale             sql.NullTime    `json:"last_sale"`
	ProvisionedBalance   sql.NullFloat64 `json:"provisioned_balance"`
	ReservedBalance      sql.NullFloat64 `json:"reserved_balance"`
	AvailableBalance     sql.NullFloat64 `json:"available_balance"`
	CreatedAt            sql.NullTime    `json:"created_at"`
	Status               sql.NullBool    `json:"status"`
	BrandActivity        sql.NullString  `json:"brand_activity"`
	BrandLogoURL         sql.NullString  `json:"brand_logo_url"`
}
