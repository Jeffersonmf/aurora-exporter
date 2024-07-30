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

func FactOrdersExporter(ctx context.Context) {

	const REDSHIFT_QUERY = `SELECT pk_id, connector_tag, workspace_id, source_id, connection_id, order_id, customer_id, total, shipping, discount, cancel_reason, currency, internal_id, "source", details, status, active, processed_at, created_at, updated_at, deleted_at, cancelled_at
	FROM ecommerce.fact_orders;`

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("FactOrdersExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("FactOrdersExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteFactOrders(ctx)
		if err != nil {
			util.Sugar.Errorf("FactOrdersExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.FactOrder{}

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

			arg := hotdata.InsertFactOrderParams{
				PkID:         sql.NullInt32{Valid: true, Int32: int32(itemRow.PKID.Int64)},
				ConnectorTag: sql.NullString{Valid: true, String: itemRow.ConnectorTag.String},
				WorkspaceID:  itemRow.WorkspaceID,
				SourceID:     itemRow.SourceID,
				ConnectionID: itemRow.ConnectionID,
				OrderID:      sql.NullString{Valid: true, String: itemRow.OrderID.String},
				CustomerID:   sql.NullString{Valid: true, String: itemRow.CustomerID.String},
				Total:        sql.NullString{Valid: true, String: fmt.Sprint(itemRow.Total.Float64)},
				Shipping:     sql.NullString{Valid: true, String: fmt.Sprint(itemRow.Shipping.Float64)},
				Discount:     sql.NullString{Valid: true, String: fmt.Sprint(itemRow.Discount.Float64)},
				CancelReason: sql.NullString{Valid: true, String: itemRow.CancelReason.String},
				Currency:     sql.NullString{Valid: true, String: itemRow.Currency.String},
				InternalID:   sql.NullInt32{Valid: true, Int32: int32(itemRow.InternalID.Int64)},
				Source:       sql.NullString{Valid: true, String: itemRow.Source.String},
				Details:      sql.NullString{Valid: true, String: itemRow.Details.String},
				Status:       sql.NullString{Valid: true, String: itemRow.Status.String},
				Active:       sql.NullBool{Valid: true, Bool: itemRow.Active.Bool},
				ProcessedAt:  sql.NullTime{Valid: true, Time: itemRow.ProcessedAt.Time},
				CreatedAt:    sql.NullTime{Valid: true, Time: itemRow.CreatedAt.Time},
				UpdatedAt:    sql.NullTime{Valid: true, Time: itemRow.UpdatedAt.Time},
				DeletedAt:    sql.NullTime{Valid: true, Time: itemRow.DeletedAt.Time},
				CancelledAt:  sql.NullTime{Valid: true, Time: itemRow.CancelledAt.Time},
			}

			err = postgreshelper.Queries.InsertFactOrder(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("FactOrdersExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.ecommerce.fact_orders]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
