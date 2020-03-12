package models


// M<>M
type AccountUser struct {
	AccountId uint	`json:"accountId" gorm:"type:int;index;not null;"`
	UserId uint	`json:"userId" gorm:"type:int;index;not null;"`
	RoleId uint	`json:"roleId" gorm:"type:int;not null;"`

	User User	`json:"-"  gorm:"preload:true"`
	Account Account	`json:"-" gorm:"preload:true"`
	Role Role	`json:"-" gorm:"preload:true"`
}

func (AccountUser) PgSqlCreate() {
	// 1. Создаем таблицу и настройки в pgSql
	db.DropTableIfExists(&AccountUser{})
	db.CreateTable(&AccountUser{})

	db.Exec("ALTER TABLE account_users \n    ADD CONSTRAINT account_users_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT account_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT account_users_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\ncreate unique index uix_account_users_account_id_user_id_role_id ON account_users (account_id,user_id,role_id);\n")

}

// Установить имя таблицы AccountUser's как `account_users`
func (AccountUser) TableName() string {
	return "account_users"
}

func (aUser *AccountUser) update (input interface{}) error {

	// выбираем те поля, что можно обновить
	return db.Model(aUser).Where("account_id = ? AND user_id = ?", aUser.AccountId, aUser.UserId).
		Select("AccountId", "UserId", "RoleId").
		Update(input).First(aUser).Error
}


