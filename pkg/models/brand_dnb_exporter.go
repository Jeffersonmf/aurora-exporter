package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"

	"brandlovrs.exporter/db/hotdata"
	"brandlovrs.exporter/internal/core"
	"brandlovrs.exporter/internal/database/postgreshelper"
	"brandlovrs.exporter/internal/database/redshifthelper"
	"brandlovrs.exporter/pkg/entities"
	"brandlovrs.exporter/pkg/util"
	workermanager "github.com/Jeffersonmf/go-workers/v3/pkg/worker_manager"
	"go.uber.org/zap"
)

func BrandDnbExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from datahub.rds_mysql.brand_dnb bd"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("BrandDnbExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("BrandDnbExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.DeleteBrandDnbTable()
		if err != nil {
			util.Sugar.Errorf("BrandDnbExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			branddnb := entities.BrandDnb{}

			s := reflect.ValueOf(&branddnb).Elem()
			numCols := s.NumField()
			columns := make([]interface{}, numCols)
			for i := 0; i < numCols; i++ {
				field := s.Field(i)
				columns[i] = field.Addr().Interface()
			}

			err := rows.Scan(columns...)
			if err != nil {
				log.Fatal(err)
			}

			arg := hotdata.InsertBrandDnbToHotdataParams{
				ID:                      branddnb.ID,
				Cnpj:                    branddnb.CNPJ,
				Name:                    branddnb.Name,
				MinimumBalance:          sql.NullString{Valid: true, String: fmt.Sprint(branddnb.MinimumBalance.Float64)},
				Url:                     branddnb.URL,
				BrandBannerUrl:          branddnb.BrandBannerURL,
				FkIDAcceptRule:          sql.NullInt32{Valid: true, Int32: branddnb.FkIDAcceptRule.Int32},
				About:                   branddnb.About,
				CreatedAt:               sql.NullTime{Valid: true, Time: branddnb.CreatedAt.Time},
				ChurnDate:               sql.NullTime{Valid: true, Time: branddnb.ChurnDate.Time},
				ChurnFinishedAt:         sql.NullTime{Valid: true, Time: branddnb.ChurnFinishedAt.Time},
				Isactive:                sql.NullBool{Valid: true, Bool: branddnb.IsActive.Bool},
				DatahubWorkspaceID:      branddnb.DatahubWorkspaceID,
				BrandLogoUrl:            branddnb.BrandLogoURL,
				BrandBannerUrlThumb:     branddnb.BrandBannerURLThumb,
				NumberOfEmployeesID:     branddnb.NumberOfEmployeesID,
				AverageMonthlyRevenueID: branddnb.AverageMonthlyRevenueID,
				FkIDBrandStatus:         branddnb.FkIDBrandStatus,
				ForceApplyTakeRate:      sql.NullBool{Valid: true, Bool: branddnb.ForceApplyTakeRate.Bool},
				AlwaysApplyTakeRate:     sql.NullBool{Valid: true, Bool: branddnb.AlwaysApplyTakeRate.Bool},
				CompanyName:             branddnb.CompanyName}

			_, err = postgreshelper.AddNewBrandDnb(arg)
			if err != nil {
				util.Sugar.Errorf("BrandDnbExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.brand_dnb]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
