package entities

import (
	"database/sql"

	"github.com/google/uuid"
)

type Workspace struct {
	ID           uuid.NullUUID  `json:"id"`
	CustomerID   uuid.NullUUID  `json:"customer_id"`
	Email        sql.NullString `json:"email"`
	Name         sql.NullString `json:"name"`
	ReferenceID  sql.NullString `json:"reference_id"`
	IntegratorID sql.NullInt64  `json:"integrator_id"`
	Active       sql.NullBool   `json:"active"`
	CreatedAt    sql.NullTime   `json:"created_at"`
	UpdatedAt    sql.NullTime   `json:"updated_at"`
	Type         sql.NullString `json:"type"`
}
