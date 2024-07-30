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

func ColabUserExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from datahub.rds_mysql.colab_user;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("ColabUser::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("ColabUser::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.DeleteColabUserTable()
		if err != nil {
			util.Sugar.Errorf("ColabUser::DeleteColabUserTable() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			var itemRow = entities.ColabUser{}

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

			arg := hotdata.InsertColabUserToHotdataParams{
				ID:                     itemRow.ID.Int32,
				Name:                   sql.NullString{Valid: true, String: util.FilterInvalidUTF8(itemRow.Name.String)},
				Email:                  itemRow.Email,
				Password:               itemRow.Password,
				Isactive:               itemRow.IsActive,
				FkIDColabUserInvite:    itemRow.FkIDColabUserInvite,
				Termsuse:               itemRow.TermsUse,
				Emailconfirmationtoken: itemRow.EmailConfirmationToken,
				Token:                  sql.NullString{Valid: true, String: util.FilterInvalidUTF8(itemRow.Token.String)},
				CreatedAt:              itemRow.CreatedAt,
				UpdatedAt:              itemRow.UpdatedAt,
				Username:               itemRow.Username,
				Avatar:                 itemRow.Avatar,
				ThumbAvatar:            itemRow.ThumbAvatar,
				TokenPush:              itemRow.TokenPush,
				CreatorSizeID:          itemRow.CreatorSizeID,
				LastSocialReportSync:   itemRow.LastSocialReportSync,
				Activity:               itemRow.Activity,
				FirstAccess:            itemRow.FirstAccess}

			_, err = postgreshelper.AddNewColabuser(arg)
			if err != nil {
				util.Sugar.Errorf("ColabUser::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[datahub.rds_mysql.colab_user]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
