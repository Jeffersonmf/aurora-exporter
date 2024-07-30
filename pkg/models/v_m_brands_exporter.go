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

func VMBrandsExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from brandlovrs.v_m_brands;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("VMBrandsExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("VMBrandsExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteVMBrands(ctx)
		if err != nil {
			util.Sugar.Errorf("VMBrandsExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.VMBrands{}

			s := reflect.ValueOf(&itemRow).Elem()
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

			arg := hotdata.InsertVMBrandsParams{
				BrandID:              sql.NullInt32{Valid: true, Int32: int32(itemRow.BrandID.Int64)},
				Brand:                sql.NullString{Valid: true, String: itemRow.Brand.String},
				Cnpj:                 sql.NullString{Valid: true, String: itemRow.CNPJ.String},
				CompanyName:          sql.NullString{Valid: true, String: itemRow.CompanyName.String},
				Interest:             sql.NullString{Valid: true, String: itemRow.Interest.String},
				Challenges:           sql.NullInt32{Valid: true, Int32: int32(itemRow.Challenges.Int64)},
				ActiveChallenges:     sql.NullInt32{Valid: true, Int32: int32(itemRow.ActiveChallenges.Int64)},
				TotalAmbassadors:     sql.NullInt32{Valid: true, Int32: int32(itemRow.TotalAmbassadors.Int64)},
				AmbassadorsWithSales: sql.NullInt32{Valid: true, Int32: int32(itemRow.AmbassadorsWithSales.Int64)},
				TotalSales:           sql.NullString{Valid: true, String: fmt.Sprint(itemRow.TotalSales.Float64)},
				BrandlovrsRate:       sql.NullString{Valid: true, String: fmt.Sprint(itemRow.BrandlovrsRate.Float64)},
				Gmv:                  sql.NullString{Valid: true, String: fmt.Sprint(itemRow.GMV.Float64)},
				Gmv7:                 sql.NullString{Valid: true, String: fmt.Sprint(itemRow.GMV7.Float64)},
				Gmv15:                sql.NullString{Valid: true, String: fmt.Sprint(itemRow.GMV15.Float64)},
				Gmv30:                sql.NullString{Valid: true, String: fmt.Sprint(itemRow.GMV30.Float64)},
				Gmv60:                sql.NullString{Valid: true, String: fmt.Sprint(itemRow.GMV60.Float64)},
				Gmv90:                sql.NullString{Valid: true, String: fmt.Sprint(itemRow.GMV90.Float64)},
				LastSale:             sql.NullTime{Valid: true, Time: itemRow.LastSale.Time},
				ProvisionedBalance:   sql.NullString{Valid: true, String: fmt.Sprint(itemRow.ProvisionedBalance.Float64)},
				ReservedBalance:      sql.NullString{Valid: true, String: fmt.Sprint(itemRow.ReservedBalance.Float64)},
				AvailableBalance:     sql.NullString{Valid: true, String: fmt.Sprint(itemRow.AvailableBalance.Float64)},
				CreatedAt:            sql.NullTime{Valid: true, Time: itemRow.CreatedAt.Time},
				Status:               sql.NullBool{Valid: true, Bool: itemRow.Status.Bool},
				BrandActivity:        sql.NullString{Valid: true, String: itemRow.Brand.String},
				BrandLogoUrl:         sql.NullString{Valid: true, String: itemRow.BrandLogoURL.String}}

			err = postgreshelper.Queries.InsertVMBrands(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("VMBrandsExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[brandlovrs.v_m_brands]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
