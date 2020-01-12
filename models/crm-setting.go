package models

import (
	"time"
)

// this is CRM settings. If json - public, else - private
type CrmSetting struct {
	UserRegistrationAllow bool `json:"-" gorm:"user_registration_allow;default:true"`
	UserRegistrationInviteOnly bool `json:"user_registration_invite_only" gorm:"user_registration_invite_only;default:true"`

	CreatedAt 	time.Time `json:"-"`
	UpdatedAt 	time.Time `json:"-"`
	//DeletedAt 	*time.Time `json:"-" db:"deleted_at"`
}

func (settings *CrmSetting) Create() error {
	return db.Create(settings).Error
}
// Берет по первому ID
func (CrmSetting) Get () (*CrmSetting, error) {
	settings := &CrmSetting{}
	err := db.First(settings).Error;

	return settings, err
}