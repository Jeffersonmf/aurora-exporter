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

func VBrandTotalBalanceExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from brandlovrs.v_brand_total_balance;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("VBrandTotalBalanceExporter:RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("VBrandTotalBalanceExporter:QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteVBrandTotalBalance(ctx)
		if err != nil {
			util.Sugar.Errorf("VBrandTotalBalanceExporter:DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.VBrandTotalBalance{}

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

			arg := hotdata.InsertVBrandTotalBalanceParams{
				BrandID:               itemRow.BrandID,
				Name:                  itemRow.Name,
				WalletCredits:         itemRow.WalletCredits,
				Provisioned:           itemRow.Provisioned,
				ReservedDebit:         itemRow.ReservedDebit,
				Reserved:              itemRow.Reserved,
				EngRewardsPaidOld:     itemRow.EngRewardsPaidOld,
				EngRewardsPaidCurrent: itemRow.EngRewardsPaidCurrent,
				TotalRandomPayments:   itemRow.TotalRandomPayments,
				AdjustDebits:          itemRow.AdjustDebits,
				Available:             itemRow.Available,
			}

			err = postgreshelper.Queries.InsertVBrandTotalBalance(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("VBrandTotalBalanceExporter:writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[brandlovrs.v_brand_total_balance]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
