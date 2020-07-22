package models

import (
	"errors"
	"github.com/fatih/structs"
	"github.com/nkokorev/crm-go/utils"
)

// M<>M
type AccountUser struct {
	AccountID uint	`json:"accountID" gorm:"type:int;index;not null;"`
	UserID uint	`json:"userID" gorm:"type:int;index;not null;"`
	RoleID uint	`json:"roleID" gorm:"type:int;not null;"`

	User    User    `json:"-"  gorm:"preload:true"`

	//User `json:"user"  gorm:"preload:true"`
	Account Account `json:"-" gorm:"preload:true"`
	//Role    Role    `json:"role" gorm:"preload:true"`
	Role    Role    `json:"role" gorm:"preload:true"`
	// Role    Role    `json:"role"  gorm:"many2many:account_users;preload"`
	// Role []Role `json:"role" gorm:"many2many:account_users;preload"`
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

func (aUser AccountUser) create () (*AccountUser, error) {

	var e utils.Error

	// more Validate!
	if aUser.AccountID < 1 {
		e.AddErrors("accountID", "Необходимо указать принадлежность к аккаунту")
	}
	if aUser.UserID < 1 {
		e.AddErrors("userID", "Необходимо указать принадлежность к пользователю")
	}
	if aUser.RoleID < 1 {
		e.AddErrors("roleID", "Необходимо указать роль пользователя")
	}
	
	if  !(Account{}).Exist(aUser.AccountID) {
		return nil, errors.New("Аккаунт, в рамках которого создается пользователь, не существует!")
	}
	if  !(User{ID: aUser.UserID}).Exist() {
		return nil, errors.New("Аккаунт, в рамках которого создается пользователь, не существует!")
	}
	if  !(Role{ID: aUser.RoleID}).Exist() {
		return nil, errors.New("Аккаунт, в рамках которого создается пользователь, не существует!")
	}

	if e.HasErrors() {
		e.Message = "Не верно сформированные данные пользователя для добавления в аккаунт!"
		return nil, e
	}

	if err := db.Model(&aUser).Create(&aUser).Preload("User").Preload("Account").Preload("Role").Find(&aUser).Error; err != nil {
		return nil, err
	}

	return &aUser, nil
}

func (aUser *AccountUser) update (input interface{}) error {
	return db.Model(AccountUser{}).Where("account_id = ? AND user_id = ?", aUser.AccountID, aUser.UserID).
		Updates(input).Preload("Account").Preload("User").Preload("Role").First(aUser).Error
}


func (aUser *AccountUser) delete() error {
	
	if aUser.AccountID < 1 || aUser.UserID < 1 || aUser.RoleID <1 {
		return errors.New("Не возможно удалить пользователя, т.к. не верные входящие данные")
	}
	return db.Model(AccountUser{}).Where("account_id = ? AND user_id = ?", aUser.AccountID, aUser.UserID).Delete(aUser).Error
}

func (AccountUser) SelectArrayWithoutBigObject() []string {
	fields := structs.Names(&AccountUser{}) //.(map[string]string)
	fields = utils.RemoveKey(fields, "Account")
	fields = utils.RemoveKey(fields, "User")
	fields = utils.RemoveKey(fields, "Role")
	return utils.ToLowerSnakeCaseArr(fields)
}
