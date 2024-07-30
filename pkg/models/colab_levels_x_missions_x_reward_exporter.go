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

func ColabLevelsXColabMissionsXColabRewardExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from rds_mysql.colab_levels_x_missions_x_reward;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabLevelsXColabMissionsXColabRewardExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("ColabLevelsXColabMissionsXColabRewardExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteColabLevelsXMissionsXReward(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabLevelsXColabMissionsXColabRewardExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.ColabLevelsXColabMissionsXColabReward{}

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

			arg := hotdata.InsertColabLevelsXMissionsXRewardParams{
				FkIDColabMissions:   sql.NullInt32{Valid: true, Int32: int32(itemRow.FkIDColabMissions.Int64)},
				FkIDColabLevels:     sql.NullInt32{Valid: true, Int32: int32(itemRow.FkIDColabLevels.Int64)},
				FkIDColabTypeReward: sql.NullInt32{Valid: true, Int32: int32(itemRow.FkIDColabTypeReward.Int64)},
				GrossValue:          sql.NullString{Valid: true, String: fmt.Sprint(itemRow.GrossValue.Float64)},
				Value:               sql.NullString{Valid: true, String: fmt.Sprint(itemRow.Value.Float64)}}

			err = postgreshelper.Queries.InsertColabLevelsXMissionsXReward(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("ColabLevelsXColabMissionsXColabRewardExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.colab_levels_x_missions_x_reward]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
