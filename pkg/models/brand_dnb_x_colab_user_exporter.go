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

func BrandDnbXColabUserExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from rds_mysql.brand_dnb_x_colab_user;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("BrandDnbXColabUserExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("BrandDnbXColabUserExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteBrandDnbXColabUser(ctx)
		if err != nil {
			util.Sugar.Errorf("BrandDnbXColabUserExporter::DeleteBrandDnbXColabUser() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.BrandColabUser{}

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

			arg := hotdata.InsertBrandDnbXColabUserParams{
				ID:                            itemRow.ID,
				FkIDColabUser:                 itemRow.ColabUserID,
				FkIDBrandDnb:                  itemRow.BrandDNBID,
				FkIDBrandDnbXInfluencerInvite: itemRow.BrandDNBInfluencerInviteID,
				FlagApplyTakeRate:             itemRow.ApplyTakeRateFlag,
			}

			err = postgreshelper.Queries.InsertBrandDnbXColabUser(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("BrandDnbXColabUserExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[rds_mysql.brand_dnb_x_colab_user]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
