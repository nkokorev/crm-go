package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	e "github.com/nkokorev/crm-go/errors"
	"reflect"
)

type AccountUser struct {
	ID        	uint `gorm:"primary_key;unique_index;" json:"-"`
	UserID 		uint // belong to user
	AccountID 	uint // belong to account
	RoleID   	uint `json:"role_id"`
	ApiKeys		[]ApiKey `json:"-"` // ???
}

// Вспомогательная функция получения пользователя ассоциированного с пользователем
// todo: надо бы переименовать эту функцию
func (aUser *AccountUser) GetByUserAccountID(user_id, account_id uint) error {

	if err := base.GetDB().First(aUser,"user_id = ? AND account_id = ?", user_id, account_id).Error; err != nil {
		return err
	}
	return nil
}

// устанавливает новую роль пользователю. Запрет на изменение роли owner user.
func (aUser *AccountUser) SetRole(role *Role) error {

	currentRole := Role{}
	err := base.GetDB().Model(aUser).Related(&currentRole).Error;
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	// нельзя менять роль владельца аккаунта на какую-либо еще
	if currentRole.Tag == "owner" {
		return e.RoleChangeOwnerRoleFailed
	}

	if err := base.GetDB().Model(aUser).Update("RoleID", role.ID).Error; err != nil {
		return err
	}
	return nil
}

// получить роль пользователя
func (aUser *AccountUser) GetRole(role *Role) error {

	// Найдем роль пользователя (ее может не быть)
	if err := base.GetDB().Model(aUser).Related(role).Error;err != nil {
		return err
	}
	return nil
}

// ниже вспомогательные функции простановки системных ролей пользователям
func (aUser *AccountUser) SetRoleOwner() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.FindRoleByTag("owner");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := aUser.SetRole(&role); err != nil {
		return err
	}

	return nil
}
func (aUser *AccountUser) SetRoleAdmin() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.FindRoleByTag("admin");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := aUser.SetRole(&role);err != nil {
		return err
	}
	return nil
}
func (aUser *AccountUser) SetRoleManager() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.FindRoleByTag("manager");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := aUser.SetRole(&role);err != nil {
		return err
	}
	return nil
}
func (aUser *AccountUser) SetRoleMarketer() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.FindRoleByTag("marketer");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := aUser.SetRole(&role);err != nil {
		return err
	}
	return nil
}
func (aUser *AccountUser) SetRoleAuthor() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.FindRoleByTag("author");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := aUser.SetRole(&role);err != nil {
		return err
	}
	return nil
}
func (aUser *AccountUser) SetRoleViewer() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.FindRoleByTag("viewer");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := aUser.SetRole(&role);err != nil {
		return err
	}
	return nil
}

// Проверяет, является ли aUser владельцем связанного аккаунта
func (aUser *AccountUser) IsOwner() bool{

	// найдем аккаунт пользователя
	account := Account{}
	if err := base.GetDB().Model(aUser).Related(&account); err != nil {
		return false
	}

	// ID владелец аккаунта совпадает с ID аккаунта связанного с пользователем?
	if account.ID == aUser.AccountID {
		return true
	}

	return false
}

// Проверяем права доступа для конкретного пользователя.
func (aUser *AccountUser) CheckPermission(permission_code uint) bool {

	// 1. Получаем роль пользователя
	role := Role{}
	if err := aUser.GetRole(&role); err != nil {
		// у пользователя может не быть роли в аккаунте, в этом случае он безправный
		return false
	}

	// 2. Если пользователь имеет роль владельца аккаунта (для API своя проверка) - сразу даем зеленый свет - это ускоряет работу.
	if role.Tag == "owner" {
		return true
	}

	// Если роль отличная от owner, то сделаем прямой запрос в БД через функцию Роли
	if role.hasPermission(permission_code) {
		return true
	}

	// не нашли доказательств наличия права
	return false
}

// функция проверяет существенная ли модель
func (aUser *AccountUser) isExists() bool {
	if reflect.TypeOf(aUser.ID).String() != "uint" || aUser.ID < 1 || base.GetDB().First(&AccountUser{}, aUser.ID).RecordNotFound() {
		return false
	}
	return true
}
// обратная к isExists функция
func (aUser *AccountUser) isNotExists() bool {
	return aUser.isExists()
}