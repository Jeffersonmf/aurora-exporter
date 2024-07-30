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

func VMPostMetricsExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from social_metrics.v_m_post_metrics;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("VMPostMetricsExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("VMPostMetricsExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeletePostMetricsToHotdata(ctx)
		if err != nil {
			util.Sugar.Errorf("VMPostMetricsExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.PostMetrics{}

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

			arg := hotdata.InsertPostMetricsToHotdataParams{
				PostID:          itemRow.PostID.Int64,
				SocialNetworkID: sql.NullInt32{Valid: true, Int32: itemRow.SocialNetworkID.Int32},
				Url:             sql.NullString{Valid: true, String: itemRow.URL.String},
				ProfileID:       sql.NullInt64{Valid: true, Int64: itemRow.ProfileID.Int64},
				LastUpdated:     sql.NullTime{Valid: true, Time: itemRow.LastUpdated.Time},
				Likes:           sql.NullInt32{Valid: true, Int32: int32(itemRow.Likes.Int64)},
				Comments:        sql.NullInt32{Valid: true, Int32: int32(itemRow.Comments.Int64)},
				Plays:           sql.NullInt32{Valid: true, Int32: int32(itemRow.Plays.Int64)},
				Shares:          sql.NullInt32{Valid: true, Int32: int32(itemRow.Shares.Int64)}}

			err = postgreshelper.Queries.InsertPostMetricsToHotdata(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("VMPostMetricsExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[social_metrics.v_m_post_metrics]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
