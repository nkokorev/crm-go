package models

import (
	"time"
)

// this is CRM settings. If json - public, else - private
type CrmSetting struct {
	Id uint `json:"-"`

	// Глобальные настройки
	ApiEnabled 		bool `json:"api_enabled" gorm:"default:true;not null"` // влючен ли API интерфейс для аккаунтов
	AppUiApiEnabled bool `json:"app_ui_api_enabled" gorm:"default:true;not null"` // Включен ли APP UI-API интерфейс (через https://app.ratuscrm.com/ui-api/)
	UiApiEnabled 	bool `json:"uiApiEnabled" gorm:"default:true;not null"` // Включен ли публичный UI-API интерфейс (через https://ui.api.ratuscrm.com)

	ApiDisabledMessage 		string `json:"api_disabled_message" gorm:"type:varchar(255);"`
	UiApiDisabledMessage 	string `json:"ui_api_disabled_message" gorm:"type:varchar(255);"`
	AppUiApiDisabledMessage string `json:"app_ui_api_disabled_message" gorm:"type:varchar(255);"`

	// SMTPPrivateAPIKey string `json:"-" gorm:"type:varchar(255);default:'cd00e0c60b26be77e32a943bd5768a19-65b08458-9049e45c'"` // MailGunKey private api key

	CreatedAt 	time.Time `json:"-"`
	UpdatedAt 	time.Time `json:"-"`
	//DeletedAt 	*time.Time `json:"-" db:"deleted_at"`
}

func (CrmSetting) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	if db.Migrator().HasTable(&CrmSetting{}) { return }
	db.Migrator().CreateTable(&CrmSetting{})

	settings := &CrmSetting{
		ApiEnabled: true,
		UiApiEnabled: true,
		AppUiApiEnabled: true,
		ApiDisabledMessage: "Sorry, the server is under maintenance.",
		UiApiDisabledMessage: "Sorry, the server is under maintenance.",
		AppUiApiDisabledMessage: "Из-за работ на сервере интерфейс временно отключен.",
		// SMTPPrivateAPIKey: "cd00e0c60b26be77e32a943bd5768a19-65b08458-9049e45c",
	}
	db.Create(&settings)
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