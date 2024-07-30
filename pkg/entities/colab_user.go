package entities

import (
	"database/sql"
)

type ColabUser struct {
	ID                     sql.NullInt32  `json:"id"`
	Name                   sql.NullString `json:"name"`
	Email                  sql.NullString `json:"email"`
	Password               sql.NullString `json:"password"`
	IsActive               sql.NullBool   `json:"isActive"`
	FkIDColabUserInvite    sql.NullInt32  `json:"fk_id_colab_user_invite"`
	TermsUse               sql.NullBool   `json:"termsUse"`
	EmailConfirmationToken sql.NullString `json:"emailConfirmationToken"`
	Token                  sql.NullString `json:"token"`
	CreatedAt              sql.NullTime   `json:"created_at"`
	UpdatedAt              sql.NullTime   `json:"updated_at"`
	Username               sql.NullString `json:"username"`
	Avatar                 sql.NullString `json:"avatar"`
	ThumbAvatar            sql.NullString `json:"thumbAvatar"`
	TokenPush              sql.NullString `json:"tokenPush"`
	CreatorSizeID          sql.NullInt32  `json:"creatorSizeID"`
	LastSocialReportSync   sql.NullTime   `json:"lastSocialReportSync"`
	Activity               sql.NullString `json:"activity"`
	FirstAccess            sql.NullBool   `json:"first_access"`
}
