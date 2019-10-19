package models

import (
	"fmt"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
	"reflect"
)

/**
* Ролевая модель определяет объединяет в группу права (permissions) через role_permissions
* При пересечении ролей, определяющей будет роль с наиболее широкими правами (модель perm.Status || perm-2.Status || ...)
* Однако, при определении прав доступа будет приоритетным прямое назначение: aUsers | aUser <> permissions
 */
type Role struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID string `json:"hash_id" gorm:"type:varchar(10);unique_index;"`
	AccountID uint `json:"account_id"` // belong to account account owner, foreign_key
	Name string `json:"name" gorm:"size:255"` // название роли в системе: Администратор / Менеджер / Оператор / Кладовщик / Маркетолог
	Description string `json:"description" gorm:"size:255;"` // Описание роли: 'Роль для новых администраторов склада...'
	AUsers []AccountUser `json:"user_id" gorm:"many2many:account_user_roles;"` // у одного пользователя может быть несколько ролей
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"` // одна роль имеет много прав (permissions)
}

// создает роль и ассоциирует ее с правами (values[0] = Permission || [] Permissions)
func (role *Role) Create(account Account, values... interface{}) (error u.Error) {
	db := database.GetDB()

	role.HashID, error = CreateHashID(role)
	if error.HasErrors() {
		error.Message = t.Trans(t.RoleFailedToCreate)
		return
	}

	err := db.Model(&account).Association("Roles").Append(role).Error
	if err != nil {
		error.Message = t.Trans(t.RoleFailedToCreate)
		return
	}

	if (len(values) > 0) {
		role.PermissionAdd(values[0])
	}

	return
}

// удаляет роль (в контексте из аккаунта, т.к. каждая роль привязана строго к 1 аккаунту)
func (role Role) Delete() (error u.Error) {

	if reflect.TypeOf(role.ID).String() != "uint" {
		error.Message = t.Trans(t.RoleDeletionError)
		return
	}

	if err := database.GetDB().Unscoped().Delete(&role).Error; err != nil {
		error.Message = t.Trans(t.RoleDeletionError)
		return
	}

	return
}

// связывает роль и разрешения: Permission, []Permissions
func (role Role) PermissionAdd(values interface{}) (error u.Error) {

	if err := database.GetDB().Model(role).Association("Permissions").Append(values).Error; err != nil {
		error.Message = t.Trans(t.RoleFailedAddPermissions)
		return
	}
	return
}

// удаляет разрешение(я) для роли Permission, []Permissions
func (role Role) PermissionRemove(values interface{}) (error u.Error) {

	if err := database.GetDB().Model(role).Association("Permissions").Delete(values).Error; err != nil {
		error.Message = t.Trans(t.RoleFailedRemovePermissions)
		return
	}
	return
}

// добавляет роль пользователю
func (aUser AccountUser) RoleAdd(role Role) (error u.Error) {
	err := database.GetDB().Model(&aUser).Association("Roles").Append(&role).Error
	if err != nil {
		fmt.Println("Неудалось добавить пользователю роль", err)
		return
	}
	return
}

// добавляет роль пользователю
func (aUser AccountUser) RoleRemove(role Role) (error u.Error) {
	err := database.GetDB().Model(&aUser).Association("Roles").Delete(&role).Error
	if err != nil {
		fmt.Println("Неудалось добавить пользователю роль", err)
		return
	}
	return
}

