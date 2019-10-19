package models

import (
	"fmt"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
)


const (
	PermissionCreateAccount 		 = 10 // Право на создание нового аккаунта (= true)
	PermissionStoreListing           = 101 // Доступ к списку складов
	PermissionStoreEditing           = 102 // Редактирование данных склада
	PermissionStoreCreating          = 103 // Создание склада
	PermissionStoreDeleting          = 104 // Удаление склада
)

var permissions = []Permission{

	// Аккаунт: Роли 101-109
	{Name: "Управление ролями", Tag:"account", CodeName: "PermissionRoleManagement", Code: 101, Description: "Возможность управлять ролями (создавать, редактировать и удалять)."},
	{Name: "Управление API-ключами", Tag:"account", CodeName: "PermissionAPIManagement", Code: 102, Description: "Возможность управлять API-ключами (создавать, редактировать и удалять)."},

	// Пользователи: 2хх |
	{Name: "Просмотр пользователей", Tag:"user", CodeName: "PermissionUserListing", Code: 201, Description: "Доступ к чтению списка пользователей, их правам и данным самих пользователей."},
	{Name: "Редактирование пользователя", Tag:"user", CodeName: "PermissionUserEditing", Code: 202, Description: "Возможность редактировать данные пользователя в аккаунте, включая его права (кроме владельца аккаунта)."},
	{Name: "Добавление пользователей", Tag:"user", CodeName: "PermissionUserAppend", Code: 203, Description: "Возможность приглашать пользователей в аккаунт."},
	{Name: "Удаление пользователей", Tag:"user", CodeName: "PermissionUserDeleting", Code: 204, Description: "Возможность исключать пользователей из аккаунта."},

	// Склады, товары, услуги: Склады 401-410
	{Name: "Просмотр складов", Tag:"store", CodeName: "PermissionStoreListing", Code: 401, Description: "Доступ к чтению списка складов и их внутренних данных."},
	{Name: "Редактирование склада", Tag:"store", CodeName: "PermissionStoreEditing", Code: 402, Description: "Возможность редактировать данные уже созданного склада."},
	{Name: "Создание склада", Tag:"store", CodeName: "PermissionStoreCreating", Code: 403, Description: "Возможность создать склад."},
	{Name: "Удаление склада", Tag:"store", CodeName: "PermissionStoreDeleting", Code: 404, Description: "Возможность удалить склад со всеми его данными."},


}

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




// Создание нового правила доступа (сугубо внутренняя функция)
func (permission *Permission) Create() (error u.Error) {

	err := database.GetDB().Create(permission).Error
	if err != nil {
		error.Message = t.Trans(t.PermissionFailedToCreate)
		return
	}
	return
}

// разворачивает базовые разрешения для всех пользователей
func PermissionSeeding()  {
	for _, v := range permissions {
		err := v.Create()
		if err.HasErrors() {
			fmt.Println("Cant create Permissions")
		}
	}
}

// Add Permission To AccountUser
func (permission *Permission) AddToUser(aUser *AccountUser) (error u.Error) {
	if err := database.GetDB().Model(aUser).Association("Permissions").Append(permission).Error; err != nil {
		error.Message = "Cant add permission to User"
	}
	return
}


// Add Permission To AccountUser
func (aUser *AccountUser) PermissionAdd(i interface{}) (error u.Error) {

	permission := &Permission{}

	switch i.(type) {

	case int, uint:
		if database.GetDB().First(&permission, "code = ?", i.(int)).RecordNotFound() {
			error.Message = "Cant find permission code: " + fmt.Sprint(i);
		}
	case *Permission:
		permission = i.(*Permission)
	default:
		error.Message = "Cant add permission to User"
		return
	}

	error = permission.AddToUser(aUser)

	return
}

// Remove Permission From AccountUser
func (permission *Permission) RemoveFromUser(aUser *AccountUser) (error u.Error) {
	if err := database.GetDB().Model(aUser).Association("Permissions").Delete(permission).Error; err != nil {
		error.Message = "Cant add permission to User"
	}
	return
}


// Remove Permission From AccountUser
func (aUser *AccountUser) PermissionRemove(i interface{}) (error u.Error) {

	permission := &Permission{}

	switch i.(type) {

	case int, uint:
		if database.GetDB().First(&permission, "code = ?", i.(int)).RecordNotFound() {
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








func (user *User) PermissionCheck(CheckedPermissions uint) (status bool) {

	// 1. Проверяем есть ли прямое назначение нужной нам роли (персональная роль), т.к. она перебивает остальные


	// 2. Ищем нужный пермишенс через Роли и проверяем его статус

	status = false // default
	return true
}





