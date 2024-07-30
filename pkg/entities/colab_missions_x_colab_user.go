package entities

import (
	"database/sql"
)

type ColabMissionsXColabUser struct {
	ID                sql.NullInt32 `json:"id"`
	FKIDColabMissions sql.NullInt32 `json:"fk_id_colab_missions"`
	FKIDColabUser     sql.NullInt32 `json:"fk_id_colab_user"`
	Status            sql.NullBool  `json:"status"`
	FlagApplyTakeRate sql.NullBool  `json:"flag_apply_take_rate"`
}
