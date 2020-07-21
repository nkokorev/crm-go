package models

import (
	"errors"
	"github.com/fatih/structs"
	"github.com/nkokorev/crm-go/utils"
)

// M<>M
type AccountUser struct {
	AccountId uint	`json:"accountId" gorm:"type:int;index;not null;"`
	UserId uint	`json:"userId" gorm:"type:int;index;not null;"`
	RoleId uint	`json:"roleId" gorm:"type:int;not null;"`

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
	if aUser.AccountId < 1 {
		e.AddErrors("accountId", "Необходимо указать принадлежность к аккаунту")
	}
	if aUser.UserId < 1 {
		e.AddErrors("userId", "Необходимо указать принадлежность к пользователю")
	}
	if aUser.RoleId < 1 {
		e.AddErrors("roleId", "Необходимо указать роль пользователя")
	}
	
	if  !(Account{}).Exist(aUser.AccountId) {
		return nil, errors.New("Аккаунт, в рамках которого создается пользователь, не существует!")
	}
	if  !(User{ID: aUser.UserId}).Exist() {
		return nil, errors.New("Аккаунт, в рамках которого создается пользователь, не существует!")
	}
	if  !(Role{ID: aUser.RoleId}).Exist() {
		return nil, errors.New("Аккаунт, в рамках которого создается пользователь, не существует!")
	}

	if e.HasErrors() {
		e.Message = "Не верно сформированные данные пользователя для добавления в аккаунт!"
		return nil, e
	}

	if err := db.Model(&aUser).Create(&aUser).Preload("User").Preload("Account").Preload("Role").Find(&aUser).Error; err != nil {
		return nil, err
	}

	// fmt.Println("### UserId:  ", aUser.Role)

	return &aUser, nil
}

func (aUser *AccountUser) update (input interface{}) error {
	return db.Model(AccountUser{}).Where("account_id = ? AND user_id = ?", aUser.AccountId, aUser.UserId).
		Updates(input).Preload("Account").Preload("User").Preload("Role").First(aUser).Error
}

/*func (aUser *AccountUser) save () error {
	return db.Model(&AccountUser{}).Save(aUser).Preload("Account").Preload("User").Preload("Role").First(aUser).Error
}*/

func (aUser *AccountUser) delete() error {
	
	if aUser.AccountId < 1 || aUser.UserId < 1 || aUser.RoleId <1 {
		return errors.New("Не возможно удалить пользователя, т.к. не верные входящие данные")
	}
	return db.Model(AccountUser{}).Where("account_id = ? AND user_id = ?", aUser.AccountId, aUser.UserId).Delete(aUser).Error
}

func (AccountUser) SelectArrayWithoutBigObject() []string {
	fields := structs.Names(&AccountUser{}) //.(map[string]string)
	fields = utils.RemoveKey(fields, "Account")
	fields = utils.RemoveKey(fields, "User")
	fields = utils.RemoveKey(fields, "Role")
	return utils.ToLowerSnakeCaseArr(fields)
}
