package entities

import "database/sql"

type ColabLevelsXColabMissionsXColabReward struct {
	FkIDColabMissions   sql.NullInt64   `json:"fkIdColabMissions"`
	FkIDColabLevels     sql.NullInt64   `json:"fkIdColabLevels"`
	FkIDColabTypeReward sql.NullInt64   `json:"fkIdColabTypeReward"`
	GrossValue          sql.NullFloat64 `json:"grossValue"`
	Value               sql.NullFloat64 `json:"value"`
}
