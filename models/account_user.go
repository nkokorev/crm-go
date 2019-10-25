package models

import (
	"fmt"
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
	Permissions []Permission `json:"permissions" gorm:"many2many:account_user_permissions;"`

	ApiKeys		[]ApiKey `json:"-"`
}

// устанавливает новую роль пользователю. (временный) Запрет на изменение роли owner user.
func (aUser *AccountUser) SetNewRole(role *Role) error {

	currentRole := Role{}
	err := base.GetDB().Model(aUser).Related(&currentRole).Error;
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		fmt.Println("Hash is not", aUser, currentRole)
		return err
	}

	if currentRole.Tag == "owner" {
		return e.RoleChangeOwnerRoleFailed
	}

	if err := base.GetDB().Model(aUser).Update("RoleID", role.ID).Error; err != nil {
		return err
	}
	return nil
}

// Вспомогательная функция получения пользователя ассоциированного с пользователем
func (aUser *AccountUser) GetAccountUser(user_id, account_id uint) error {
	err := base.GetDB().First(aUser,"user_id = ? AND account_id = ?", user_id, account_id).Error
	if err != nil {
		return err
	}
	return nil
}

func (aUser *AccountUser) SetOwnerRole() error {
	role := Role{}
	err := base.GetDB().First(&role, "tag = 'owner'").Error
	if err != nil {
		return err
	}

	if err := aUser.SetNewRole(&role); err != nil {
		return err
	}
	return nil
}
func (aUser *AccountUser) SetAdminRole() error {
	role := Role{}
	err := base.GetDB().First(&role, "tag = 'admin'").Error
	if err != nil {
		return err
	}

	if err := aUser.SetNewRole(&role);err != nil {
		return err
	}
	return nil
}
func (aUser *AccountUser) SetManagerRole() error {
	role := Role{}
	err := base.GetDB().First(&role, "tag = 'manager'").Error
	if err != nil {
		return err
	}

	if err := aUser.SetNewRole(&role);err != nil {
		return err
	}
	return nil
}
func (aUser *AccountUser) SetMarketerRole() error {
	role := Role{}
	err := base.GetDB().First(&role, "tag = 'marketer'").Error
	if err != nil {
		return err
	}

	if err := aUser.SetNewRole(&role);err != nil {
		return err
	}
	return nil
}
func (aUser *AccountUser) SetAuthorRole() error {
	role := Role{}
	err := base.GetDB().First(&role, "tag = 'author'").Error
	if err != nil {
		return err
	}

	if err := aUser.SetNewRole(&role);err != nil {
		return err
	}
	return nil
}
func (aUser *AccountUser) SetViewerRole() error {
	role := Role{}
	err := base.GetDB().First(&role, "tag = 'viewer'").Error
	if err != nil {
		return err
	}

	if err := aUser.SetNewRole(&role);err != nil {
		return err
	}
	return nil
}