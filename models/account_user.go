package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	e "github.com/nkokorev/crm-go/errors"
)

type AccountUser struct {
	ID        	uint `gorm:"primary_key;unique_index;" json:"-"`
	UserID 		uint // belong to user
	AccountID 	uint // belong to account
	RoleID   	uint `json:"role_id"`
	ApiKeys		[]ApiKey `json:"-"` // ???
}

// Вспомогательная функция получения пользователя ассоциированного с пользователем
func (aUser *AccountUser) GetAccountUser(user_id, account_id uint) error {
	err := base.GetDB().First(aUser,"user_id = ? AND account_id = ?", user_id, account_id).Error
	if err != nil {
		return err
	}
	return nil
}

// устанавливает новую роль пользователю. (временный) Запрет на изменение роли owner user.
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
