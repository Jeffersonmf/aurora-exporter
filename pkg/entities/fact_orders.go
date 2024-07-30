package entities

import (
	"database/sql"

	"github.com/google/uuid"
)

type FactOrder struct {
	PKID         sql.NullInt64   `json:"pk_id"`
	ConnectorTag sql.NullString  `json:"connector_tag"`
	WorkspaceID  uuid.NullUUID   `json:"workspace_id"`
	SourceID     uuid.NullUUID   `json:"source_id"`
	ConnectionID uuid.NullUUID   `json:"connection_id"`
	OrderID      sql.NullString  `json:"order_id"`
	CustomerID   sql.NullString  `json:"customer_id"`
	Total        sql.NullFloat64 `json:"total"`
	Shipping     sql.NullFloat64 `json:"shipping"`
	Discount     sql.NullFloat64 `json:"discount"`
	CancelReason sql.NullString  `json:"cancel_reason"`
	Currency     sql.NullString  `json:"currency"`
	InternalID   sql.NullInt64   `json:"internal_id"`
	Source       sql.NullString  `json:"source"`
	Details      sql.NullString  `json:"details"`
	Status       sql.NullString  `json:"status"`
	Active       sql.NullBool    `json:"active"`
	ProcessedAt  sql.NullTime    `json:"processed_at"`
	CreatedAt    sql.NullTime    `json:"created_at"`
	UpdatedAt    sql.NullTime    `json:"updated_at"`
	DeletedAt    sql.NullTime    `json:"deleted_at"`
	CancelledAt  sql.NullTime    `json:"cancelled_at"`
}
