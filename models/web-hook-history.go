package models

import "log"

type WebHookHistory struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	WebHookId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Status 		bool 	`json:"enabled" gorm:"type:bool;default:true"` // успешен ли вызо
}

func (WebHookHistory) PgSqlCreate() {
	db.Migrator().CreateTable(&WebHookHistory{})
	// db.Model(&WebHookHistory{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&WebHookHistory{}).AddForeignKey("web_hook_id", "web_hooks(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE web_hook_histories " +
		"ADD CONSTRAINT web_hook_histories_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT web_hook_histories_web_hook_id_fkey FOREIGN KEY (web_hook_id) REFERENCES web_hooks(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

