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

func ColabUserDataExporter(ctx context.Context) {

	const REDSHIFT_QUERY = `SELECT id, value, fk_id_colab_type_label, fk_id_colab_user FROM rds_mysql.colab_user_data;`

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabUserDataExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("ColabUserDataExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteColabUserData(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabUserDataExporter::DeleteColabUserData() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.ColabUserData{}

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

			arg := hotdata.InsertColabUserDataParams{
				ID:                 sql.NullInt32{Valid: true, Int32: int32(itemRow.ID.Int64)},
				Value:              sql.NullString{Valid: true, String: itemRow.Value.String},
				FkIDColabTypeLabel: sql.NullInt32{Valid: true, Int32: int32(itemRow.FkIDColabTypeLabel.Int64)},
				FkIDColabUser:      sql.NullInt32{Valid: true, Int32: int32(itemRow.FkIDColabTypeLabel.Int64)},
			}

			err = postgreshelper.Queries.InsertColabUserData(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("ColabUserDataExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.colab_user_data]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
