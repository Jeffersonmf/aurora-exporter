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

func VMBrandWalletHistoryExporter(ctx context.Context) {

	const REDSHIFT_QUERY = `SELECT brand_id, brand_name, 
	mission_id, mission_name, challenge_type, value, order_id, transaction_date, source_balance, target_balance, transaction_type, details
	 FROM brandlovrs.v_m_brand_wallet_history;`

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("BrandWalletHistoryExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("BrandWalletHistoryExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteBrandWalletHistory(ctx)
		if err != nil {
			util.Sugar.Errorf("BrandWalletHistoryExporter::DeleteBrandWalletHistory() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.BrandWalletHistory{}

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

			arg := hotdata.InsertBrandWalletHistoryParams{
				BrandID:         sql.NullInt32{Valid: true, Int32: int32(itemRow.BrandID.Int64)},
				BrandName:       sql.NullString{Valid: true, String: itemRow.BrandName.String},
				MissionID:       sql.NullInt32{Valid: true, Int32: int32(itemRow.MissionID.Int64)},
				MissionName:     sql.NullString{Valid: true, String: itemRow.MissionName.String},
				ChallengeType:   sql.NullString{Valid: true, String: itemRow.ChallengeType.String},
				Value:           sql.NullString{Valid: true, String: fmt.Sprint(itemRow.Value.Float64)},
				OrderID:         sql.NullString{Valid: true, String: itemRow.OrderID.String},
				TransactionDate: sql.NullTime{Valid: true, Time: itemRow.TransactionDate.Time},
				SourceBalance:   sql.NullString{Valid: true, String: itemRow.SourceBalance.String},
				TargetBalance:   sql.NullString{Valid: true, String: itemRow.TargetBalance.String},
				TransactionType: sql.NullString{Valid: true, String: itemRow.TransactionType.String},
				Details:         sql.NullString{Valid: true, String: itemRow.Details.String},
			}

			err = postgreshelper.Queries.InsertBrandWalletHistory(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("BrandWalletHistoryExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.v_m_brand_wallet_history]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
