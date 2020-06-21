package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

type WebHook struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // обрабатывать ли вебхук
	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя вебхука
	URL 		string 	`json:"url" gorm:"type:varchar(255);"` // вызов, который совершается
}


func (WebHook) PgSqlCreate() {
	db.CreateTable(&WebHook{})
	db.Model(&WebHook{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

func (wh *WebHook) BeforeCreate(scope *gorm.Scope) error {
	wh.ID = 0
	return nil
}

func (WebHook) TableName() string {
	return "web_hooks"
}

// ######### CRUD Functions ############
func (wh WebHook) create() (*WebHook, error) {
	var whNew = wh
	err := db.Create(&whNew).First(&whNew).Error
	return &whNew, err
}

func (WebHook) get(id uint) (*WebHook, error) {

	wh := WebHook{}

	if err := db.First(&wh, id).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}

func (WebHook) getListByAccount(accountId uint) ([]WebHook, error) {

	whs := make([]WebHook,0)

	err := db.Find(&whs, "accountId = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return whs, nil

}

func (wh *WebHook) update(input interface{}) error {
	return db.Model(wh).Omit("id", "account_id").Update(structs.Map(input)).Error

}

func (wh WebHook) delete () error {
	return db.Model(WebHook{}).Where("id = ?", wh.ID).Delete(wh).Error
}
// ######### END CRUD Functions ############

func (account Account) GetWebHook(id uint) (*WebHook, error) {

	wh, err := WebHook{}.get(id)
	if err != nil {
		return nil, err
	}

	if wh.AccountID != account.ID {
		return nil, utils.Error{Message: "WebHook принадлежит другому аккаунту"}
	}

	return wh, nil
}
func (account Account) GetWebHooks() ([]WebHook, error) {
	return WebHook{}.getListByAccount(account.ID)
}