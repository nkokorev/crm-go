package models

import (
	"fmt"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	e "github.com/nkokorev/crm-go/errors"
	u "github.com/nkokorev/crm-go/utils"
	"reflect"
)

const (
	RoleOverallAdmin 		 = 100 // Право на создание нового аккаунта (= true)
	RoleOverallManager           = 101 // Доступ к списку складов
	RoleOverallMarketer           = 102 // Редактирование данных склада
)

var (
	// Доступ ко всем данным и функционалу аккаунта
	permissionsOwner = []int{
		PermissionUserListing,
		PermissionUserEditing,
		PermissionUserAppend,
		PermissionUserDeleting,

		PermissionStoreListing,
		PermissionStoreEditing,
		PermissionStoreCreating,
		PermissionStoreDeleting,
		PermissionProductListing,
		PermissionProductEditing,
		PermissionProductCreating,
		PermissionProductDeleting,

		PermissionAPIManagement,
	}

	// Как и владелец, но не может менять владельца аккаунта и биллинговую информацию
	permissionsAdmin = []int{
		PermissionUserListing,
		PermissionUserEditing,
		PermissionUserAppend,
		PermissionUserDeleting,

		PermissionStoreListing,
		PermissionStoreEditing,
		PermissionStoreCreating,
		PermissionStoreDeleting,
		PermissionProductListing,
		PermissionProductEditing,
		PermissionProductCreating,
		PermissionProductDeleting,

		PermissionAPIManagement,
	}

	// Не может управлять пользователями, смотреть и изменять биллинговую и системную информацию
	permissionsManager = []int{
		PermissionUserListing,
		PermissionUserEditing,
		PermissionUserAppend,
		PermissionUserDeleting,

		PermissionStoreListing,
		PermissionStoreEditing,
		PermissionStoreCreating,
		PermissionStoreDeleting,
		PermissionProductListing,
		PermissionProductEditing,
		PermissionProductCreating,
		PermissionProductDeleting,

		PermissionAPIManagement,
	}

	// Может читать отчеты, доступ к аналитике, может управлять разделом маркетинга + доступ к редактуре текстов как у автора БЕЗ рецензирования.
	permissionsMarketer = []int{
		PermissionUserListing,
		PermissionUserEditing,
		PermissionUserAppend,
		PermissionUserDeleting,

		PermissionStoreListing,
		PermissionStoreEditing,
		PermissionStoreCreating,
		PermissionStoreDeleting,
		PermissionProductListing,
		PermissionProductEditing,
		PermissionProductCreating,
		PermissionProductDeleting,

		PermissionAPIManagement,
	}

	// может создавать и редактировать описания товаров, статей, письма и т.д., но не может запускать их в продакшен без рецензирования вышестоящего (спец.роль принятия правок).
	permissionsAuthor = []int{
		PermissionUserListing,
		PermissionUserEditing,
		PermissionUserAppend,
		PermissionUserDeleting,

		PermissionStoreListing,
		PermissionStoreEditing,
		PermissionStoreCreating,
		PermissionStoreDeleting,
		PermissionProductListing,
		PermissionProductEditing,
		PermissionProductCreating,
		PermissionProductDeleting,

		PermissionAPIManagement,
	}

	// не может вносить изменения (никакие!!!). Может смотреть отчеты, товары, склады, письма и т.д.
	permissionsViewer = []int{
		PermissionUserListing,
		PermissionUserEditing,
		PermissionUserAppend,
		PermissionUserDeleting,

		PermissionStoreListing,
		PermissionStoreEditing,
		PermissionStoreCreating,
		PermissionStoreDeleting,
		PermissionProductListing,
		PermissionProductEditing,
		PermissionProductCreating,
		PermissionProductDeleting,

		PermissionAPIManagement,
	}

)

// Список ролей в системе. Каждая роль имеет список прав (permissions).
// Некоторые аккаунты могут заводить собственные роли (adv accounting).
type Role struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID string `json:"hash_id" gorm:"type:varchar(10);unique_index;"`
	Tag 		string `json:"tag" gorm:"size:255;unique;" ` // admin, manager, marketer...
	AccountID uint `json:"-"` // belong to account account owner, foreign_key <= реализация в будущем!!
	System 		bool `json:"system" gorm:"default:false"` // дефолтная ли роль или нет
	Name string `json:"name" gorm:"size:255"` // название роли в системе: Администратор / Менеджер / Оператор / Кладовщик / Маркетолог
	Description string `json:"description" gorm:"size:255;"` // Описание роли: 'Роль для новых администраторов склада...'
	AUsers 		[]AccountUser `json:"-"`
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

// системная функция: удаляет роль
func (role *Role) Delete() error {

	if reflect.TypeOf(role.ID).String() != "uint" {
		return e.RoleDeletionError
	}

	// проверяем, есть ли привязанные пользователи
	if base.GetDB().Model(role).Association("AUsers").Count() > 0 {
		return e.RoleDeletedFailedHasUsers
	}

	if err := base.GetDB().Unscoped().Delete(&role).Error; err != nil {
		return err
	}
	return nil
}

// устанавливает для роли ТОЛЬКО переданные разрешения, остальные разрешения удаляются
func (role *Role) SetPermissions(codes []int) error {
	for i, v := range codes {
		fmt.Printf("SetPermissions: i: %v, v: %v \r\n", i,v)
	}
	return nil
}

// добавляет роли переданные разрешения, остальные разрешения остаются без изменений
func (role *Role) AppendPermissions(codes []int) error {
	for i, v := range codes {
		fmt.Printf("AppendPermissions: i: %v, v: %v \r\n", i,v)
	}
	return nil
}
// убирает у роли переданные разрешения
func (role *Role) RemovePermissions(codes []int) error {
	for i, v := range codes {
		fmt.Printf("RemovePermissions: i: %v, v: %v \r\n", i,v)
	}
	return nil
}

// вспомогательная функция поиска роли по тегу
func (role *Role) GetRoleByTag(tag string) error {
	if err := base.GetDB().First(&role, "tag = ?", tag).Error; err != nil {
		return err
	}
	return nil
}

// связывает роли и разрешения: []Permissions
func (role *Role) AppendPermissionsOLD(permissions []Permission) error {

	if err := base.GetDB().Model(role).Association("Permissions").Append(&permissions).Error; err != nil {
		return err
	}
	return nil
}

// развязывает роли и разрешения
func (role *Role) RemovePermissionsOLD(permissions []Permission) error {

	if err := base.GetDB().Model(role).Association("Permissions").Delete(&permissions).Error; err != nil {
		return err
	}
	return nil
}

// Назначает роль пользователю
func (role *Role) AppendUser(aUser *AccountUser) error {
	return aUser.SetNewRole(role)
}


// todo реализовать сидерские функции базового наполнения crm таблиц данных

// системная функция для установки прав администратора
// todo реализовать функционал
func (role *Role) setPermissionsOwner() error {
	if err := role.SetPermissions(permissionsOwner); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setPermissionsAdmin() error {
	if err := role.SetPermissions(permissionsAdmin); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setPermissionsManager() error {
	if err := role.SetPermissions(permissionsManager); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setPermissionsMarketer() error {
	if err := role.SetPermissions(permissionsMarketer); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setPermissionsAuthor() error {
	if err := role.SetPermissions(permissionsAuthor); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setPermissionsViewer() error {
	if err := role.SetPermissions(permissionsViewer); err != nil {
		return nil
	}
	return nil
}



// добавляет к роли стандартные права администратора
func (role Role) AppendAdminPermissions() (error u.Error) {
	// todo дописать функцию
	return
}



