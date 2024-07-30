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

func MissionXCouponInfoExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from rds_mysql.mission_x_coupon_info"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("MissionXCouponInfoExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("MissionXCouponInfoExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteMissionXCouponInfo(ctx)
		if err != nil {
			util.Sugar.Errorf("MissionXCouponInfoExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.MissionXCouponInfo{}

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

			arg := hotdata.InsertMissionXCouponInfoParams{
				ID:                     sql.NullInt32{Valid: true, Int32: int32(itemRow.ID)},
				ExpiresAt:              itemRow.ExpiresAt,
				TotalQuantity:          itemRow.TotalQuantity,
				BrandPrefix:            itemRow.BrandPrefix,
				Value:                  sql.NullString{Valid: true, String: fmt.Sprint(itemRow.Value.Float64)},
				FkIDCouponTypeDiscount: itemRow.FkIDCouponTypeDiscount,
				FkIDColabMission:       itemRow.FkIDColabMission,
				MaxUsePerCreator:       itemRow.MaxUsePerCreator}

			err = postgreshelper.Queries.InsertMissionXCouponInfo(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("MissionXCouponInfoExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.mission_x_coupon_info]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
