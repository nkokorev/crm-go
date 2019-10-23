package models

import (
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
)

type AccountUser struct {
	ID        	uint `gorm:"primary_key;unique_index;" json:"-"`
	UserID 		uint // belong to user
	AccountID 	uint // belong to account
	Permissions []Permission `json:"permissions" gorm:"many2many:account_user_permissions;"`
	Roles   	[]Role `json:"roles" gorm:"many2many:account_user_roles;"`
	ApiKeys		[]ApiKey `json:"-"`
}

// добавляет роль пользователю
func (aUser *AccountUser) AppendRole(role Role) (error u.Error) {
	err := base.GetDB().Model(&aUser).Association("Roles").Append(&role).Error
	if err != nil {
		error.Message = t.Trans(t.UserFailedAddRole)
		return
	}
	return
}

// удаляет роль у пользователя
func (aUser *AccountUser) RemoveRole(role Role) (error u.Error) {
	err := base.GetDB().Model(&aUser).Association("Roles").Delete(&role).Error
	if err != nil {
		error.Message = t.Trans(t.UserFailedRemoveRole)
		return
	}
	return
}

// Вспомогательная функция получения пользователя ассоциированного с пользователем
func (aUser *AccountUser) GetAccountUser(user_id, account_id uint) error {
	err := base.GetDB().Model(&aUser).First(aUser, "account_id = ? AND user_id = ?", user_id, account_id).Error
	if err != nil {
		return err
	}
	return nil
}

