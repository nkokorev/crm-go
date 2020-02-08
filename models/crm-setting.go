package models

import (
	"errors"
	"time"
)

// this is CRM settings. If json - public, else - private
type CrmSetting struct {
	ID uint `json:"-"`

	// Глобальные настройки
	ApiEnabled bool `json:"apiEnabled" gorm:"default:true;not null"` // влючен ли API интерфейс
	AppUiApiEnabled bool `json:"appUiApiEnabled" gorm:"default:true;not null"` // Включен ли APP UI-API интерфейс (через https://app.ratuscrm.com/ui-api/)
	UiApiEnabled bool `json:"uiApiEnabled" gorm:"default:true;not null"` // Включен ли публичный UI-API интерфейс (через https://ui.api.ratuscrm.com)

	ApiDisabledMessage string `json:"apiDisableMessage" gorm:"type:varchar(255);"`
	UiApiDisabledMessage string `json:"uiApiDisableMessage" gorm:"type:varchar(255);"`
	AppUiApiDisabledMessage string `json:"appUiApiDisableMessage" gorm:"type:varchar(255);"`

	CreatedAt 	time.Time `json:"-"`
	UpdatedAt 	time.Time `json:"-"`
	//DeletedAt 	*time.Time `json:"-" db:"deleted_at"`
}

// внутренняя чит-фукнция для создания системных настроек
func CreateCrmSettings() (*CrmSetting, error) {

	if !db.Model(&CrmSetting{}).First(&CrmSetting{}, "id = 1").RecordNotFound() {
		return nil, errors.New("Настройки CRM уже загружены!")
	}

	settings := &CrmSetting{
		ApiEnabled: true,
		UiApiEnabled: true,
		AppUiApiEnabled: true,
		ApiDisabledMessage: "Sorry, the server is under maintenance.",
		UiApiDisabledMessage: "Sorry, the server is under maintenance.",
		AppUiApiDisabledMessage: "Из-за работ на сервере интерфейс временно отключен.",
	}

	err := db.Create(&settings).Error

	return settings, err
}

// Берет первую строку т.е. должна быть единственная запись
func GetCrmSettings () (*CrmSetting, error) {
	settings := &CrmSetting{}
	err := db.First(settings).Error

	return settings, err
}

// сохраняет текущее состояние настроек в структуруе
func (settings *CrmSetting) Save () error {
	return db.Model(settings).Omit("id", "created_at", "updated_at").Save(settings).First(settings).Error
}