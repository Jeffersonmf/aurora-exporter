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

func VMsnPostHashtagsExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from social_metrics.v_m_sn_post_hashtags;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("VMsnPostHashtagsExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("VMsnPostHashtagsExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeletePostMetricsToHotdata(ctx)
		if err != nil {
			util.Sugar.Errorf("VMsnPostHashtagsExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.VMsnPostHashtags{}

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

			arg := hotdata.InsertSNPostHashtagsToHotdataParams{
				UserID:    sql.NullInt32{Valid: true, Int32: int32(itemRow.UserID.Int64)},
				Name:      sql.NullString{Valid: true, String: itemRow.Name.String},
				ProfileID: sql.NullInt64{Valid: true, Int64: itemRow.ProfileID.Int64},
				PostID:    sql.NullInt64{Valid: true, Int64: itemRow.PostID.Int64},
				CreatedAt: sql.NullTime{Valid: true, Time: itemRow.CreatedAt.Time},
				MediaID:   sql.NullInt32{Valid: true, Int32: itemRow.MediaID.Int32},
				HashtagBl: sql.NullString{Valid: true, String: itemRow.HashtagBl.String},
			}

			_, err = postgreshelper.Queries.InsertSNPostHashtagsToHotdata(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("VMsnPostHashtagsExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[social_metrics.v_m_sn_post_hashtags]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
