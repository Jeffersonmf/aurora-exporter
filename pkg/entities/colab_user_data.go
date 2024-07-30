package entities

import "database/sql"

type ColabUserData struct {
	ID                 sql.NullInt64  `json:"id"`
	Value              sql.NullString `json:"value"`
	FkIDColabTypeLabel sql.NullInt64  `json:"fk_id_colab_type_label"`
	FkIDColabUser      sql.NullInt64  `json:"fk_id_colab_user"`
}
