package models

import (
	"context"
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

func ColabCouponExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from datahub.rds_mysql.colab_coupon;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabCouponExporter:RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("ColabCouponExporter:QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteColabCoupon(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabCouponExporter:DeleteColabUserable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			var itemRow = entities.ColabCoupon{}

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

			arg := hotdata.InsertColabCouponParams{
				ID:                     itemRow.ID,
				CreatedAt:              itemRow.CreatedAt,
				InfluencerCouponPrefix: itemRow.InfluencerCouponPrefix,
				Quantity:               itemRow.Quantity,
				FkIDColabUser:          itemRow.FkIDColabUser,
				FkIDMissionXCouponInfo: itemRow.FkIDMissionXCouponInfo,
				EcommerceCouponID:      itemRow.EcommerceCouponID,
				AmountOfUse:            itemRow.AmountOfUse,
				LastUse:                itemRow.LastUse,
			}

			err = postgreshelper.Queries.InsertColabCoupon(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("ColabCouponExporter:writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.colab_coupon]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
