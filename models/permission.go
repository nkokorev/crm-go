package models

import (
	"errors"
	"fmt"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	u "github.com/nkokorev/crm-go/utils"
)

const (
	PermissionCreateAccount 		 = 10 // Право на создание нового аккаунта (= true)
	PermissionStoreListing           = 101 // Доступ к списку складов
	PermissionStoreEditing           = 102 // Редактирование данных склада
	PermissionStoreCreating          = 103 // Создание склада
	PermissionStoreDeleting          = 104 // Удаление склада
)

// Список возможных разрешений в системе - один для всех аккаунтов.
type Permission struct {
	ID        	uint `json:"-" gorm:"primary_key;unique_index;"`
	Name	 	string `json:"name" gorm:"size:255;unique;"` // Просмотр товаров / Редактирование товара / Создание товара / Удаление товара - для удобства
	Tag 		string `json:"tag" gorm:"size:255"` // store, order, leads, contacts...
	CodeName 	string `json:"name" gorm:"size:255;unique_index;"` // PermissionStoreListing, PermissionStoreEditing, ...
	Code 		uint `json:"type" gorm:"unique_index;"`
	Description string `json:"description" gorm:"size:255;"` // Описание права: 'Право на редактирование товара, изображений и т.д.'
	//Seeding 	bool `json:"-"` // входит ли в дефолтное разрешение пользователя
	AUsers   	[]AccountUser `json:"-" gorm:"many2many:account_user_permissions;"`
	Roles  		[]Role `json:"-" gorm:"many2many:role_permissions;"`
	ApiKeys 	[]ApiKey `json:"-" gorm:"many2many:api_key_permissions;"`
}


// Создание нового правила доступа (сугубо системная функция)
func (permission *Permission) Create() error {
	if err := base.GetDB().Create(permission).Error; err != nil {
		return err
	}
	return nil
}


// Add Permission To AccountUser
func (permission *Permission) AddToUser(aUser *AccountUser) error {
	if err := base.GetDB().Model(aUser).Association("Permissions").Append(permission).Error; err != nil {
		return err
	}
	return nil
}

// Add Permission To AccountUser
func (aUser *AccountUser) AppendPermission(i interface{}) error {

	permission := &Permission{}

	switch i.(type) {

	case int, uint:
		if base.GetDB().First(&permission, "code = ?", i.(int)).RecordNotFound() {
			return errors.New("Cant find permission code: " + fmt.Sprint(i))
		}
	case *Permission:
		permission = i.(*Permission)
	default:
		return errors.New("Cant add permission to User")
	}

	err := permission.AddToUser(aUser)
	if err != nil {
		return err
	}
	return nil
}

// Remove Permission From AccountUser
func (permission *Permission) RemoveFromUser(aUser *AccountUser) (error u.Error) {
	if err := base.GetDB().Model(aUser).Association("Permissions").Delete(permission).Error; err != nil {
		error.Message = "Cant add permission to User"
	}
	return
}

// Remove Permission([]Permission) From AccountUser
func (aUser *AccountUser) RemovePermission(i interface{}) (error u.Error) {

	permission := &Permission{}

	switch i.(type) {

	case int, uint:
		if base.GetDB().First(&permission, "code = ?", i.(int)).RecordNotFound() {
			error.Message = "Cant find permission code: " + fmt.Sprint(i);
		}
	case *Permission:
		permission = i.(*Permission)
	default:
		error.Message = "Cant remove permission from User"
		return
	}

	error = permission.RemoveFromUser(aUser)

	return
}


func (aUser *AccountUser) PermissionCheck(CheckedPermissions uint) (status bool) {

	// 1. Проверяем владелец или нет ((a *AccountUser) isOwner() bool)
	// 2. Смотрим персональные права (permissions)
	// 3. Смотрим права через роли ()


	status = false // default
	return true
}





