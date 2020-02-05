package models

import (
	"time"
)

// this is CRM settings. If json - public, else - private
type CrmSetting struct {
	ID uint `json:"-"`

	// Глобальные настройки
	ApiEnabled bool `json:"apiEnabled" gorm:"default:true;not null"` // влючен ли API интерфейс
	UiApiPublicEnabled bool `json:"uiApiEnabled" gorm:"default:false;not null"` // Включен ли Public UI-API интерфейс (через https://ui.api.ratuscrm.com)
	ApiDisabledMessage string `json:"apiDisableMessage" gorm:"type:varchar(255);"`
	UiApiDisabledMessage string `json:"uiApiDisableMessage" gorm:"type:varchar(255);"`

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

func (settings *CrmSetting) Save () error {
	return db.Model(settings).Omit("id", "created_at", "updated_at").Save(settings).First(settings).Error
}

func (settings *CrmSetting) Update (input interface{}) error {
	return db.Model(settings).Omit("id", "created_at", "updated_at").Update(input).First(settings).Error
}