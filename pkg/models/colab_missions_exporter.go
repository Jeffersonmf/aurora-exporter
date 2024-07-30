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

func ColabMissionExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from rds_mysql.colab_missions;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabMissionExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("ColabMissionExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteColabMissionsToHotdata(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabMissionExporter::DeleteColabMissionsToHotdata() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.ColabMission{}

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

			arg := hotdata.InsertColabMissionsToHotdataParams{

				ID:                  itemRow.ID.Int32,
				Title:               sql.NullString{Valid: true, String: itemRow.Title.String},
				Description:         sql.NullString{Valid: true, String: itemRow.Description.String},
				FkIDColabSubmission: sql.NullInt32{Valid: true, Int32: itemRow.FKIDColabSubmission.Int32},
				FkIDBrandDnb:        sql.NullInt32{Valid: true, Int32: itemRow.FKIDBrandDnb.Int32},
				FkIDColabReward:     sql.NullInt32{Valid: true, Int32: itemRow.FKIDColabReward.Int32},
				CreatedAt:           sql.NullTime{Valid: true, Time: itemRow.CreatedAt.Time},
				ExpiredAt:           sql.NullTime{Valid: true, Time: itemRow.ExpiredAt.Time},
				Status:              sql.NullBool{Valid: true, Bool: itemRow.Status.Bool},
				PaymentDays:         sql.NullInt32{Valid: true, Int32: itemRow.PaymentDays.Int32},
				DeletedAt:           sql.NullTime{Valid: true, Time: itemRow.DeletedAt.Time},
				ShopifyPriceRuleID:  sql.NullString{Valid: true, String: itemRow.ShopifyPriceRuleID.String},
				PriceRuleSetup:      sql.NullString{Valid: true, String: itemRow.PriceRuleSetup.String},
				MinimumSubtotal:     sql.NullString{Valid: true, String: fmt.Sprint(itemRow.MinimumSubtotal.Float64)},
				Banner:              sql.NullString{Valid: true, String: itemRow.Banner.String},
				ThumbBanner:         sql.NullString{Valid: true, String: itemRow.ThumbBanner.String},
				RewardDescription:   sql.NullString{Valid: true, String: fmt.Sprint(itemRow.RewardDescription)},
				CumulativeDiscount:  sql.NullBool{Valid: true, Bool: itemRow.CumulativeDiscount.Bool},
				ChallengeStatusID:   sql.NullInt32{Valid: true, Int32: itemRow.ChallengeStatusID.Int32},
				MaxValuePerOrder:    sql.NullString{Valid: true, String: fmt.Sprint(itemRow.MaxValuePerOrder.Float64)},
				UtmSource:           sql.NullString{Valid: true, String: itemRow.UTMSource.String},
			}

			_, err = postgreshelper.Queries.InsertColabMissionsToHotdata(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("ColabMissionExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.colab_mission]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
