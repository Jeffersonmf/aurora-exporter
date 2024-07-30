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

func VMInstagramProfilesUsersExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from brandlovrs.v_m_instagram_profiles_users"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("VMInstagramProfilesUsersExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("VMInstagramProfilesUsersExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteInstagramProfilesUsersToHotdata(ctx)
		if err != nil {
			util.Sugar.Errorf("VMInstagramProfilesUsersExporter::DeleteBrandDnbTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.VMInstagramProfilesUsers{}

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

			arg := hotdata.InsertInstagramProfilesUsersToHotdataParams{
				UserID:                       sql.NullInt32{Valid: true, Int32: int32(itemRow.UserID.Int64)},
				InstagramUsername:            sql.NullString{Valid: true, String: itemRow.InstagramUsername.String},
				InstagramFullname:            sql.NullString{Valid: true, String: itemRow.InstagramFullname.String},
				InstagramID:                  sql.NullString{Valid: true, String: itemRow.InstagramID.String},
				BusinessCategory:             sql.NullString{Valid: true, String: itemRow.BusinessCategory.String},
				InstagramIgtvvideocount:      sql.NullInt32{Valid: true, Int32: int32(itemRow.InstagramIGTVVideoCount.Int64)},
				InstagramPostscount:          sql.NullInt32{Valid: true, Int32: int32(itemRow.InstagramPostsCount.Int64)},
				InstagramFollowerscount:      sql.NullInt32{Valid: true, Int32: int32(itemRow.InstagramFollowersCount.Int64)},
				InstagramFollowscount:        sql.NullInt32{Valid: true, Int32: int32(itemRow.InstagramFollowsCount.Int64)},
				InstagramDataIngestion:       sql.NullString{Valid: true, String: itemRow.InstagramDataIngestion.String},
				InstagramBiography:           sql.NullString{Valid: true, String: itemRow.InstagramBiography.String},
				InstagramPrivate:             sql.NullBool{Valid: true, Bool: itemRow.InstagramPrivate.Bool},
				InstagramVerified:            sql.NullBool{Valid: true, Bool: itemRow.InstagramVerified.Bool},
				InstagramProfileUrl:          sql.NullString{Valid: true, String: itemRow.InstagramProfileURL.String},
				InstagramProfilePictureUrl:   sql.NullString{Valid: true, String: itemRow.InstagramProfilePictureURL.String},
				InstagramProfilePictureUrlHd: sql.NullString{Valid: true, String: itemRow.InstagramProfilePictureURLHD.String},
			}

			_, err = postgreshelper.Queries.InsertInstagramProfilesUsersToHotdata(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("VMInstagramProfilesUsersExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[brandlovrs.v_m_instagram_profiles_users]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
