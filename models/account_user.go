package models

import (
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
)

type AccountUser struct {
	ID        	uint `gorm:"primary_key;unique_index;" json:"-"`
	UserID 		uint // belong to user
	AccountID 	uint // belong to account
	RoleID   	uint `json:"role_id"`
	Permissions []Permission `json:"permissions" gorm:"many2many:account_user_permissions;"`

	ApiKeys		[]ApiKey `json:"-"`
}

// добавляет роль пользователю
func (aUser *AccountUser) SetNewRole(role *Role) error {

	if err := base.GetDB().Model(aUser).Update("RoleID", role.ID).Error; err != nil {
		return err
	}
	return nil
}

// удаляет роль у пользователя
/*func (aUser *AccountUser) RemoveRole(role *Role) error {
	err := base.GetDB().Model(aUser).Association("Roles").Delete(role).Error
	if err != nil {
		return err
	}
	return nil
}*/

// Вспомогательная функция получения пользователя ассоциированного с пользователем
func (aUser *AccountUser) GetAccountUser(user_id, account_id uint) error {
	err := base.GetDB().First(aUser,"user_id = ? AND account_id = ?", user_id, account_id).Error
	if err != nil {
		return err
	}
	return nil
}



// ########### дописать и допроверить
func (aUser *AccountUser) SetOwnerRole() error {
	role := Role{}
	err := base.GetDB().First(&role, "tag = 'owner'").Error
	if err != nil {
		return err
	}

	if err := aUser.SetNewRole(&role);err != nil {
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