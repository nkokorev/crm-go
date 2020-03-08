package models

// M<>M
type AccountUser struct {
	AccountId uint	`json:"accountId" gorm:"type:int;index;not null;"`
	UserId uint	`json:"userId" gorm:"type:int;index;not null;"`
	RoleId uint	`json:"roleId" gorm:"type:int;not null;"`

	User User `json:"-"`
	Account Account	`json:"-"`
	Role Role `json:"-"`
}

func (AccountUser) PgSqlCreate() {
	// 1. Создаем таблицу и настройки в pgSql
	db.DropTableIfExists(&AccountUser{})
	db.CreateTable(&AccountUser{})

	db.Exec("ALTER TABLE account_users \n    ADD CONSTRAINT account_users_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT account_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT account_users_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\ncreate unique index uix_account_users_account_user_role_id ON account_users (account_id,user_id,role_id);\n")

}



