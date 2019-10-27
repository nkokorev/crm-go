package models

import (
	"errors"
	"fmt"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	e "github.com/nkokorev/crm-go/errors"
	u "github.com/nkokorev/crm-go/utils"
	"os"
	"reflect"
)

var roles = []Role {
	{ Name: "Владелец аккаунта",Tag:"owner", 	System: true, Description: "Доступ ко всем данным и функционалу аккаунта."},
	{ Name: "Администратор", 	Tag:"admin", 	System: true, Description: "Доступ ко всем данным и функционалу аккаунта. Не может менять владельца аккаунта."},
	{ Name: "Менеджер", 		Tag:"manager", 	System: true, Description: "Не может добавлять пользователей, не может менять биллинговую информацию."},
	{ Name: "Маркетолог", 		Tag:"marketer", System: true, Description: "Читает все клиентские данные, может изменять все что касается маркетинга, но не заказы или склады."},
	{ Name: "Автор", 			Tag:"author", 	System: true, Description: "Может создавать контент: писать статьи, письма, описания к товарам и т.д."},
	{ Name: "Наблюдатель", 		Tag:"viewer", 	System: true, Description: "The Viewer can view reports in the account"},
	{ Name: "Full Access", 		Tag:"full-access", 	Type:	"api",	System: true, Description: "Доступ ко всем функциям API"},
	{ Name: "Site Access", 		Tag:"site-access", 	Type:	"api",	System: true, Description: "Доступ к аккаунту через API, необходимый для интеграции с сайтом"},
	{ Name: "Read Access", 		Tag:"read-access", 	Type:	"api",	System: true, Description: "Доступ к чтению основной информации об аккаунте."},
}

var (
	// ### Список прав для пользователей (aUser) ### ///
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

	// ### Список прав для API ключей ### ///
	// полный доступ у API ключа
	permissionsFullAccess = []int{
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

	// список прав для интеграции с сайтом
	permissionsSiteAccess = []int{
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

	// список прав чтения статистических данных
	permissionsReadAccess = []int{
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
	Type 		string `json:"type" gorm:"size:25;default:'gui'" ` // тип роли: gui / api / any
	AccountID uint `json:"-"  gorm:"default:NULL"` // belong to account account owner, foreign_key <= реализация в будущем!!
	System 		bool `json:"system" gorm:"default:false"` // дефолтная ли роль или нет
	Name string `json:"name" gorm:"size:255"` // название роли в системе: Администратор / Менеджер / Оператор / Кладовщик / Маркетолог
	Description string `json:"description" gorm:"size:255;"` // Описание роли: 'Роль для новых администраторов склада...'
	AUsers 		[]AccountUser `json:"-"`
	APIKeys 	[]ApiKey `json:"-"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"` // одна роль имеет много прав (permissions)
}

// создает роль в системе. ?? не возможно создать роль без разрешений... же?
func (role *Role) create(codes []int) (err error) {

	// проверка на попытку создать дубль роли, которая уже была создан
	if reflect.TypeOf(role.ID).String() == "uint" {
		if role.ID > 0 && !base.GetDB().First(&Role{}, role.ID).RecordNotFound() {
			// todo need to translation
			return errors.New("Can't create new role: already crated!")
		}
	}

	if role.HashID, err = u.CreateHashID(role); err != nil {
		return err
	}

	if err = base.GetDB().Create(&role).Error; err != nil {
		return err
	}

	if err = role.SetPermissions(codes);err != nil {
		return
	}

	if err = base.GetDB().Save(role).Error; err != nil {
		return err
	}

	return
}

// удаляет роль, проверяя привязанных пользователей
func (role *Role) delete() error {

	if reflect.TypeOf(role.ID).String() != "uint" {
		return e.RoleDeletionError
	}

	// проверяем, есть ли привязанные пользователи
	if base.GetDB().Model(role).Association("AUsers").Count() > 0 {
		return e.RoleDeletedFailedHasUsers
	}

	// если все хорошо - удаляем роль с концами
	if err := base.GetDB().Unscoped().Delete(role).Error; err != nil {
		return err
	}
	return nil
}

// устанавливает для роли ТОЛЬКО переданные разрешения, остальные разрешения удаляются
func (role *Role) SetPermissions(codes []int) error {

	// ищем права с указанными кодами
	permissions := []Permission{}
	if err := base.GetDB().Find(&permissions, "code IN (?)", codes).Error; err != nil {
		return err
	}

	// назначаем права для роли
	if err := base.GetDB().Model(role).Association("Permissions").Replace(permissions).Error; err != nil {
		return nil
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
func (role *Role) FindRoleByTag(tag string) error {
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
	return aUser.SetRole(role)
}


// todo реализовать сидерские функции базового наполнения crm таблиц данных

// системные функции для установки прав для системных ролей
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
func (role *Role) setPermissionsFullAccess() error {
	if err := role.SetPermissions(permissionsFullAccess); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setPermissionsSiteAccess() error {
	if err := role.SetPermissions(permissionsSiteAccess); err != nil {
		return nil
	}
	return nil
}
func (role *Role) setPermissionsReadAccess() error {
	if err := role.SetPermissions(permissionsReadAccess); err != nil {
		return nil
	}
	return nil
}


func init() {
	// seeding can be: "" / "true" / "fresh"
	seeding := os.Getenv("seeding")
	if seeding == "true" ||  seeding == "fresh"{
		roleSeeding()
	}
}


// разворачивает базовые разрешения для всех пользователей
func roleSeeding()  {

	//base.GetDB().Unscoped().Delete(&Role{})

	// проверяем что в системе нет ролей
	if !base.GetDB().Find(&Role{}).RecordNotFound() {
		return
	}



	// создаем системные роли
	for _, v := range roles {

		// 1. Найдем необходимые пермишены
		//permI := []int{}
		switch v.Tag {
		case "owner":
			if err :=  v.create(permissionsOwner);err != nil {
				fmt.Println("Неудалось установить парава для роли owner", err.Error())
			}
		case "admin":
			if err :=  v.create(permissionsAdmin);err != nil {
				fmt.Println("Неудалось установить парава для роли admin", err.Error())
			}
		case "manager":
			if err :=  v.create(permissionsManager);err != nil {
				fmt.Println("Неудалось установить парава для роли manager", err.Error())
			}
		case "marketer":
			if err :=  v.create(permissionsMarketer);err != nil {
				fmt.Println("Неудалось установить парава для роли marketer", err.Error())
			}
		case "author":
			if err :=  v.create(permissionsAuthor);err != nil {
				fmt.Println("Неудалось установить парава для роли author", err.Error())
			}
		case "viewer":
			if err :=  v.create(permissionsViewer);err != nil {
				fmt.Println("Неудалось установить парава для роли viewer", err.Error())
			}
		case "full-access":
			if err :=  v.create(permissionsFullAccess);err != nil {
				fmt.Println("Неудалось установить парава для роли full-access", err.Error())
			}
		case "site-access":
			if err :=  v.create(permissionsSiteAccess);err != nil {
				fmt.Println("Неудалось установить парава для роли site-access", err.Error())
			}
		case "read-access":
			if err :=  v.create(permissionsReadAccess);err != nil {
				fmt.Println("Неудалось установить парава для роли read-access", err.Error())
			}
		}

		/*err := v.create()
		if err != nil {
			fmt.Println("Cant create Roles")
		}*/
	}

	// назначаем права

	/*role := Role{}
	if err := role.FindRoleByTag("owner");err != nil {
		fmt.Printf("Cant find role owner by tag: %v \r\n", err.Error())
	}
	if err := role.setPermissionsOwner();err != nil{
		fmt.Printf("Cant set permissions owner for role owner: %v", err.Error())
	}

	if err := role.FindRoleByTag("admin");err !=nil {
		fmt.Printf("Cant find role owner by tag: %v \r\n", err.Error())
	}
	if err := role.setPermissionsAdmin();err != nil{
		fmt.Printf("Cant set permissions owner for role owner: %v \r\n", err.Error())
	}*/



}



