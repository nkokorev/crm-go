package models

import "time"

// this is CRM settings. If json - public, else - private
type CrmSetting struct {
	UserRegistrationAllow bool `json:"user_registration_allow" gorm:"user_registration_allow;default:true"`
	UserRegistrationInviteOnly bool `json:"user_registration_invite_only" gorm:"user_registration_invite_only;default:true"`

	CreatedAt 	time.Time `json:"-"`
	UpdatedAt 	time.Time `json:"-"`
	//DeletedAt 	*time.Time `json:"-" db:"deleted_at"`
}

func (settings *CrmSetting) Create() error {
	return db.Create(settings).Error
}