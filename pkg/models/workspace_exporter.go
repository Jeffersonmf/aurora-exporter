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

func WorkspaceExporter(ctx context.Context) {

	const REDSHIFT_QUERY = `SELECT id, customer_id, email, "name", reference_id, integrator_id, active, created_at, updated_at, "type" FROM core.workspace;`

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("WorkspaceExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("WorkspaceExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteWorkspace(ctx)
		if err != nil {
			util.Sugar.Errorf("WorkspaceExporter::DeleteWorkspace() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.Workspace{}

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

			arg := hotdata.InsertWorkspaceParams{
				ID:           itemRow.ID,
				CustomerID:   itemRow.CustomerID,
				Email:        sql.NullString{Valid: true, String: itemRow.Email.String},
				Name:         sql.NullString{Valid: true, String: itemRow.Name.String},
				ReferenceID:  sql.NullString{Valid: true, String: itemRow.ReferenceID.String},
				IntegratorID: sql.NullInt32{Valid: true, Int32: int32(itemRow.IntegratorID.Int64)},
				Active:       sql.NullBool{Valid: true, Bool: itemRow.Active.Bool},
				CreatedAt:    sql.NullTime{Valid: true, Time: itemRow.CreatedAt.Time},
				UpdatedAt:    sql.NullTime{Valid: true, Time: itemRow.UpdatedAt.Time},
				Type:         sql.NullString{Valid: true, String: itemRow.Type.String},
			}

			err = postgreshelper.Queries.InsertWorkspace(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("WorkspaceExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.core.workspace]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
