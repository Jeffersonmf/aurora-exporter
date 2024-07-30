package entities

import (
	"database/sql"
)

type Calendar struct {
	Date      sql.NullTime
	Year      sql.NullInt32
	Month     sql.NullInt32
	Day       sql.NullInt32
	DayOfWeek sql.NullInt32
	Week      sql.NullInt32
	Quarter   sql.NullInt32
}
