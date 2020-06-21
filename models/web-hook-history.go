package models

type WebHookHistory struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	WebHookID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Status 		bool 	`json:"enabled" gorm:"type:bool;default:true"` // успешен ли вызо
}

func (WebHookHistory) PgSqlCreate() {
	db.CreateTable(&WebHookHistory{})
	db.Model(&WebHookHistory{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&WebHookHistory{}).AddForeignKey("web_hook_id", "web_hooks(id)", "CASCADE", "CASCADE")
}

