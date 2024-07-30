package entities

import (
	"database/sql"
)

type BrandDnb struct {
	ID                      int32           `json:"id"`
	Name                    sql.NullString  `json:"name"`
	URL                     sql.NullString  `json:"url"`
	BrandBannerURL          sql.NullString  `json:"brandBannerURL"`
	FkIDAcceptRule          sql.NullInt32   `json:"fkIDAcceptRule"`
	About                   sql.NullString  `json:"about"`
	CreatedAt               sql.NullTime    `json:"createdAt"`
	ChurnDate               sql.NullTime    `json:"churnDate"`
	ChurnFinishedAt         sql.NullTime    `json:"churnFinishedAt"`
	IsActive                sql.NullBool    `json:"isActive"`
	DatahubWorkspaceID      sql.NullString  `json:"datahubWorkspaceID"`
	BrandLogoURL            sql.NullString  `json:"brandLogoURL"`
	BrandBannerURLThumb     sql.NullString  `json:"brandBannerURLThumb"`
	NumberOfEmployeesID     sql.NullInt32   `json:"numberOfEmployeesID"`
	AverageMonthlyRevenueID sql.NullInt32   `json:"averageMonthlyRevenueID"`
	BrandLogoURLThumb       sql.NullString  `json:"brandLogoURLThumb"`
	MinimumBalance          sql.NullFloat64 `json:"minimumBalance"`
	CNPJ                    sql.NullString  `json:"cnpj"`
	FkIDBrandStatus         sql.NullInt32   `json:"fkIDBrandStatus"`
	ForceApplyTakeRate      sql.NullBool    `json:"forceApplyTakeRate"`
	AlwaysApplyTakeRate     sql.NullBool    `json:"alwaysApplyTakeRate"`
	CompanyName             sql.NullString  `json:"companyName"`
}
