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

func ColabMissionsXColabUserExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from rds_mysql.colab_missions_x_colab_user"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabMissionsXColabUserExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("ColabMissionsXColabUserExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteColabMissionsXColabUserToHotdata(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabMissionsXColabUserExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.ColabMissionsXColabUser{}

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

			arg := hotdata.InsertColabMissionsXColabUserToHotdataParams{
				ID:                itemRow.ID.Int32,
				FkIDColabMissions: sql.NullInt32{Valid: true, Int32: itemRow.FKIDColabMissions.Int32},
				FkIDColabUser:     sql.NullInt32{Valid: true, Int32: itemRow.FKIDColabUser.Int32},
				Status:            sql.NullBool{Valid: true, Bool: itemRow.Status.Bool},
				FlagApplyTakeRate: sql.NullBool{Valid: true, Bool: itemRow.FlagApplyTakeRate.Bool}}

			_, err = postgreshelper.Queries.InsertColabMissionsXColabUserToHotdata(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("ColabMissionsXColabUserExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.colab_missions_x_colab_user]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
