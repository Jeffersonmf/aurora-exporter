package postgreshelper

import (
	"brandlovrs.exporter/db/hotdata"
	"brandlovrs.exporter/pkg/util"
	"go.uber.org/zap"
)

func AddNewBrandDnb(brandDnbParams hotdata.InsertBrandDnbToHotdataParams) (int32, error) {

	var id int32
	id, err = Queries.InsertBrandDnbToHotdata(ctx, brandDnbParams)
	if err != nil {
		util.Sugar.Errorf(
			"postgreshelper.:",
			zap.String("msg", err.Error()),
		)
	}

	return id, err
}

func DeleteBrandDnbTable() error {

	err = Queries.DeleteBrandDnbInHotdata(ctx)
	if err != nil {
		util.Sugar.Errorf(
			"postgreshelper.:",
			zap.String("msg", err.Error()),
		)
	}

	return err
}

func AddNewColabuser(colabUserParams hotdata.InsertColabUserToHotdataParams) (int32, error) {

	var id int32

	id, err = Queries.InsertColabUserToHotdata(ctx, colabUserParams)
	if err != nil {
		util.Sugar.Errorf(
			"postgreshelper.:",
			zap.String("msg", err.Error()),
		)
	}

	return id, err
}

func DeleteColabUserTable() error {

	err = Queries.DeleteColabUserInHotdata(ctx)
	if err != nil {
		util.Sugar.Errorf(
			"postgreshelper.:",
			zap.String("msg", err.Error()),
		)
	}

	return err
}
