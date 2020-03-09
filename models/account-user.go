package models

import "errors"

// M<>M
type AccountUser struct {
	AccountId uint	`json:"accountId" gorm:"type:int;index;not null;"`
	UserId uint	`json:"userId" gorm:"type:int;index;not null;"`
	RoleId uint	`json:"roleId" gorm:"type:int;not null;"`

	User User	`sql:"-"`
	Account Account	`sql:"-"`
	Role Role	`sql:"-"`
}

func (AccountUser) PgSqlCreate() {
	// 1. Создаем таблицу и настройки в pgSql
	db.DropTableIfExists(&AccountUser{})
	db.CreateTable(&AccountUser{})

	db.Exec("ALTER TABLE account_users \n    ADD CONSTRAINT account_users_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT account_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT account_users_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\ncreate unique index uix_account_users_account_user_role_id ON account_users (account_id,user_id,role_id);\n")

}

// Установить имя таблицы AccountUser's как `account_users`
func (AccountUser) TableName() string {
	return "account_users"
}

// ### todo: нужны тесты к функциям ниже ...

func GetAccountUser(account Account, user User) (*AccountUser, error) {
	if db.NewRecord(account) || db.NewRecord(user) {
		return nil, errors.New("GetUserRole: Аккаунта или пользователя не существует!")
	}

	var aUser AccountUser

	if err := db.Table(AccountUser{}.TableName()).First(&aUser, "account_id = ? AND user_id = ?", account.ID, user.ID).Error; err != nil {
		return nil, err
	}

	return &aUser, nil
}



