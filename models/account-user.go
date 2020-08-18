package models

import (
	"database/sql"
	"errors"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

// M<>M
type AccountUser struct {
	
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;"`

	AccountId uint	`json:"accountId" gorm:"type:int;index;not null;"`
	UserId uint	`json:"userId" gorm:"type:int;index;not null;"`
	RoleId uint	`json:"roleId" gorm:"type:int;not null;"`

	User    User    `json:"-"  gorm:"preload:true"`
	Account Account `json:"-" gorm:"preload:true"`
	Role    Role    `json:"role" gorm:"preload:true"`
}

func (AccountUser) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.AutoMigrate(&AccountUser{})
	db.Model(&AccountUser{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&AccountUser{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	db.Model(&AccountUser{}).AddForeignKey("role_id", "roles(id)", "RESTRICT", "CASCADE")
	db.Exec("create unique index uix_account_users_account_id_user_id_role_id ON account_users (account_id,user_id,role_id);")
}
func (aUser *AccountUser) BeforeCreate(scope *gorm.Scope) error {

	// 1. Рассчитываем PublicId (#id заказа) внутри аккаунта
	var lastIdx sql.NullInt64

	row := db.Model(&AccountUser{}).Where("account_id = ?",  aUser.AccountId).
		Select("max(public_id)").Row()
	if row != nil {
		err := row.Scan(&lastIdx)
		if err != nil && err != gorm.ErrRecordNotFound { return err }
	} 

	aUser.PublicId = 1 + uint(lastIdx.Int64)

	return nil
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
	if  !(User{Id: aUser.UserId}).Exist() {
		return nil, errors.New("Аккаунт, в рамках которого создается пользователь, не существует!")
	}
	if  !(Role{Id: aUser.RoleId}).Exist() {
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
	return db.Model(AccountUser{}).Where("account_id = ? AND user_id = ?", aUser.AccountId, aUser.UserId).
		Updates(input).Preload("Account").Preload("User").Preload("Role").First(aUser).Error
}

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
