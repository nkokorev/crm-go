package models

import (
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
	"reflect"
)

const (
	RoleOverallAdmin 		 = 100 // Право на создание нового аккаунта (= true)
	RoleOverallManager           = 101 // Доступ к списку складов
	RoleOverallMarketer           = 102 // Редактирование данных склада
)

// Список ролей в системе. Каждая роль имеет список прав (permissions).
// Некоторые аккаунты могут заводить собственные роли (adv accounting).
type Role struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID string `json:"hash_id" gorm:"type:varchar(10);unique_index;"`
	Tag 		string `json:"tag" gorm:"size:255"` // admin, manager, marketer...
	AccountID uint `json:"-"` // belong to account account owner, foreign_key <= реализация в будущем!!
	System 		bool `json:"system"` // дефолтная ли роль или нет
	Name string `json:"name" gorm:"size:255"` // название роли в системе: Администратор / Менеджер / Оператор / Кладовщик / Маркетолог
	Description string `json:"description" gorm:"size:255;"` // Описание роли: 'Роль для новых администраторов склада...'
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"` // одна роль имеет много прав (permissions)
}

// создает роль в системе
func (role *Role) Create() (err error) {

	if role.HashID, err = u.CreateHashID(role); err != nil {
		return err
	}

	if err = base.GetDB().Create(&role).Error; err != nil {
		return err
	}
	if err = base.GetDB().Save(role).Error; err != nil {
		return err
	}

	return
}

// удаляет роль (в контексте из аккаунта, т.к. каждая роль привязана строго к 1 аккаунту)
func (role *Role) Delete() (error u.Error) {

	if reflect.TypeOf(role.ID).String() != "uint" {
		error.Message = t.Trans(t.RoleDeletionError)
		return
	}

	if err := base.GetDB().Unscoped().Delete(&role).Error; err != nil {
		error.Message = t.Trans(t.RoleDeletionError)
		return
	}

	return
}

// связывает роли и разрешения: []Permissions
func (role *Role) AppendPermissions(permissions []Permission) error {

	if err := base.GetDB().Model(role).Association("Permissions").Append(&permissions).Error; err != nil {
		return err
	}
	return nil
}

// развязывает роли и разрешения
func (role *Role) RemovePermissions(permissions []Permission) error {

	if err := base.GetDB().Model(role).Association("Permissions").Delete(&permissions).Error; err != nil {
		return err
	}
	return nil
}

// Назначает роль пользователю
func (role *Role) AppendUser(aUser AccountUser) (error u.Error) {
	return aUser.AppendRole(role)
}

// удаляет у пользователя роль
func (role *Role) RemoveUser(aUser AccountUser) (error u.Error) {
	return aUser.RemoveRole(role)
}





// todo реализовать сидерские функции базового наполнения crm таблиц данных

// системная функция для установки прав администратора
// todo реализовать функционал
func (role *Role) setOwnerPermissions() error {
	permissions := []Permission{}
	if err := role.AppendPermissions(permissions); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setAdminPermissions() error {
	permissions := []Permission{}
	if err := role.AppendPermissions(permissions); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setManagerPermissions() error {
	permissions := []Permission{}
	if err := role.AppendPermissions(permissions); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setMarketerPermissions() error {
	permissions := []Permission{}
	if err := role.AppendPermissions(permissions); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setAuthorPermissions() error {
	permissions := []Permission{}
	if err := role.AppendPermissions(permissions); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setViewerPermissions() error {
	permissions := []Permission{}
	if err := role.AppendPermissions(permissions); err != nil {
		return nil
	}
	return nil
}



// добавляет к роли стандартные права администратора
func (role Role) AppendAdminPermissions() (error u.Error) {
	// todo дописать функцию
	return
}



