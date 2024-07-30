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

func TotalCouponSalesExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from brandlovrs.v_m_total_coupon_sales;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("TotalCouponSalesExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("TotalCouponSalesExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteTotalCouponSales(ctx)
		if err != nil {
			util.Sugar.Errorf("TotalCouponSalesExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.TotalCouponSales{}

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

			arg := hotdata.InsertTotalCouponSalesParams{
				BrandID:            itemRow.BrandID,
				BrandName:          itemRow.BrandName,
				OrderID:            itemRow.OrderID,
				BlCoupon:           itemRow.BlCoupon,
				CouponCreationDate: itemRow.CouponCreationDate,
				CreatedAt:          itemRow.CreatedAt,
				UpdatedAt:          itemRow.UpdatedAt,
				Total:              sql.NullString{Valid: true, String: fmt.Sprint(itemRow.Total.Float64)},
				Discount:           sql.NullString{Valid: true, String: fmt.Sprint(itemRow.Discount.Float64)},
				Status:             itemRow.Status,
				FStatus:            itemRow.FStatus,
				MissionID:          itemRow.MissionID,
				MissionName:        itemRow.MissionName,
				TypeReward:         itemRow.TypeReward,
				IsApplyTakeRate:    itemRow.IsApplyTakeRate,
				PaymentDays:        itemRow.PaymentDays,
				RewardValue:        sql.NullString{Valid: true, String: fmt.Sprint(itemRow.RewardValue.Float64)},
				UserID:             itemRow.UserID,
				UserName:           itemRow.UserName,
				UserEmail:          itemRow.UserEmail}

			err = postgreshelper.Queries.InsertTotalCouponSales(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("TotalCouponSalesExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[brandlovrs.v_m_total_coupon_sales]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
