package models

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	u "github.com/nkokorev/crm-go/utils"
)

const (
	PermissionUserListing		= 	201
	PermissionUserEditing		= 	202
	PermissionUserAppend		= 	203
	PermissionUserDeleting		= 	204

	PermissionStoreListing		= 	401
	PermissionStoreEditing		= 	402
	PermissionStoreCreating		= 	403
	PermissionStoreDeleting		= 	404
	PermissionProductListing	= 	405
	PermissionProductEditing	= 	406
	PermissionProductCreating	= 	407
	PermissionProductDeleting	= 	408

	PermissionAPIManagement		= 	700
)

// Список возможных разрешений в системе - один для всех аккаунтов.
type Permission struct {
	ID        	uint `json:"-" gorm:"primary_key;unique_index;"`
	Name	 	string `json:"name" gorm:"size:255;unique;"` // Просмотр товаров / Редактирование товара / Создание товара / Удаление товара - для удобства
	Tag 		string `json:"tag" gorm:"size:255"` // store, order, leads, contacts...
	CodeName 	string `json:"name" gorm:"size:255;unique_index;"` // PermissionStoreListing, PermissionStoreEditing, ...
	Code 		uint `json:"type" gorm:"unique_index;"`
	Description string `json:"description" gorm:"size:255;"` // Описание права: 'Право на редактирование товара, изображений и т.д.'
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

// Ищет нужный пермишен по заданному ключу
func (permission *Permission) Find(code_int uint) error {
	// todo: дописать простой поиск по коду
	fmt.Println("Seach code: ", code_int)
	err := base.GetDB().Find(permission, "code = ?", code_int).Error;
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	return nil
}


// массовый поиск по кодам прав
func FindPermissions(values... uint) (p []Permission, err error) {
	count_v := len(values)
	for i := 0; i < count_v; i++ {
		p := Permission{}
		err = p.Find(values[i]);
		if err != nil {
			break
		}
		fmt.Println( p.Name )
	}
	return
}



// ######### ниже олдфаг функции (переписать / исключить)

// Add Permission To AccountUser (премиум функция для самых богатых)
// todo: продумать как будут добавляться прямые разрешения для пользователя и обратную функцию (которая скорее будет использована)
func (permission *Permission) AddToUser(aUser *AccountUser) error {
	if err := base.GetDB().Model(aUser).Association("Permissions").Append(permission).Error; err != nil {
		return err
	}
	return nil
}

// Add Permission To AccountUser (премиум функция для самых богатых)
// todo: еще одна функция, которая скорее должна стать основной для добавления / удаления прав у пользователя
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
// todo: доработать устаревшую фукнцию
func (permission *Permission) RemoveFromUser(aUser *AccountUser) (error u.Error) {
	if err := base.GetDB().Model(aUser).Association("Permissions").Delete(permission).Error; err != nil {
		error.Message = "Cant add permission to User"
	}
	return
}

// Remove Permission([]Permission) From AccountUser
// todo: доработать устаревшую фукнцию
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

// todo: написать функцию и тест к ней (см. тест для role <> permissions)
func (aUser *AccountUser) PermissionCheck(CheckedPermissions uint) (status bool) {

	// 1. Проверяем владелец или нет ((a *AccountUser) isOwner() bool)
	// 2. Смотрим персональные права (permissions)
	// 3. Смотрим права через роли ()


	status = false // default
	return true
}





