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

func VMInstagramProfilesBrandsExporter(ctx context.Context) {

	const REDSHIFT_QUERY = "select * from brandlovrs.v_m_instagram_profiles_brands;"

	loadDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		db, err := redshifthelper.RedshiftConnector(ctx)
		if err != nil {
			util.Sugar.Errorf("VMInstagramProfilesBrandsExporter::RedshiftConnector() -> ", zap.String("msg", err.Error()))
		}

		defer db.Close()

		rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
		if err != nil {
			util.Sugar.Errorf("VMInstagramProfilesBrandsExporter::QueryContext() -> ", zap.String("msg", err.Error()))
		}

		taskArg.SetRowsParam("rows", rows)
		return taskArg, nil
	}

	writeDataStep := func(ctx context.Context, taskArg workermanager.TaskParams) (workermanager.TaskParams, error) {

		rows, _ := taskArg.GetRowsParam("rows")

		err := postgreshelper.Queries.DeleteVMInstagramProfilesBrands(ctx)
		if err != nil {
			util.Sugar.Errorf("VMInstagramProfilesBrandsExporter::DeleteVMInstagramProfilesBrands() -> ", zap.String("msg", err.Error()))
		}

		for rows.Next() {

			itemRow := entities.VMInstagramProfileBrand{}

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

			arg := hotdata.InsertVMInstagramProfilesBrandsParams{
				BrandID:                      itemRow.BrandID,
				InstagramID:                  itemRow.InstagramID,
				InstagramUsername:            itemRow.InstagramUsername,
				InstagramFullname:            itemRow.InstagramFullname,
				InstagramIgtvvideocount:      itemRow.InstagramIGTVVideoCount,
				InstagramPostscount:          itemRow.InstagramPostsCount,
				InstagramFollowerscount:      itemRow.InstagramFollowersCount,
				InstagramFollowscount:        itemRow.InstagramFollowsCount,
				InstagramDataIngestion:       itemRow.InstagramDataIngestion,
				InstagramBiography:           itemRow.InstagramBiography,
				InstagramPrivate:             itemRow.InstagramPrivate,
				InstagramVerified:            itemRow.InstagramVerified,
				InstagramProfileUrl:          itemRow.InstagramProfileURL,
				InstagramProfilePictureUrl:   itemRow.InstagramProfilePictureURL,
				InstagramProfilePictureUrlHd: itemRow.InstagramProfilePictureURLHD,
			}

			err = postgreshelper.Queries.InsertVMInstagramProfilesBrands(ctx, arg)
			if err != nil {
				util.Sugar.Errorf("VMInstagramProfilesBrandsExporter::writeDataStep() -> ", zap.String("msg", err.Error()))
			}
		}

		println(fmt.Sprintf("Step to Write %s to Hotdata database -> Done", "[brandlovrs.v_m_instagram_profiles_brands]"))

		return taskArg, nil
	}

	core.RegisterWorker(ctx, loadDataStep, writeDataStep)
}
