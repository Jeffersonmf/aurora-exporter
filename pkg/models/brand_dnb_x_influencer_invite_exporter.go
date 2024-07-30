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

func BrandDnbXInfluencerInviteExporter(ctx context.Context) {

	const REDSHIFT_QUERY = `SELECT id, fk_id_brand_dnb, fk_id_colab_user, "acceptedAt", "createdAt", "rejectedAt", flag_who_invited, flag_apply_take_rate FROM rds_mysql.brand_dnb_x_influencer_invite;`

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("BrandDnbXInfluencerInviteExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("BrandDnbXInfluencerInviteExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteBrandDnbXInfluencerInvite(ctx)
		if err != nil {
			util.Sugar.Errorf("BrandDnbXInfluencerInviteExporter::DeleteBrandDnbXInfluencerInvite() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.BrandDnbInfluencerInvite{}

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

			arg := hotdata.InsertBrandDnbXInfluencerInviteParams{
				ID:                sql.NullInt32{Valid: true, Int32: int32(itemRow.ID.Int64)},
				FkIDBrandDnb:      sql.NullInt32{Valid: true, Int32: int32(itemRow.FkIDBrandDnb.Int64)},
				FkIDColabUser:     sql.NullInt32{Valid: true, Int32: int32(itemRow.FkIDColabUser.Int64)},
				Acceptedat:        itemRow.AcceptedAt,
				Createdat:         itemRow.CreatedAt,
				Rejectedat:        itemRow.RejectedAt,
				FlagWhoInvited:    sql.NullBool{Valid: true, Bool: itemRow.FlagWhoInvited.Bool},
				FlagApplyTakeRate: sql.NullBool{Valid: true, Bool: itemRow.FlagApplyTakeRate.Bool},
			}

			err = postgreshelper.Queries.InsertBrandDnbXInfluencerInvite(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("BrandDnbXInfluencerInviteExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.brand_dnb_x_influencer_invite]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
