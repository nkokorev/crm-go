package models

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	"os"
	"reflect"
)

// Добавил право - добавь к роли!
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
	PermissionRoleManagement	= 	701
	PermissionBillingManagement	= 	702
	PermissionAccountManagement	= 	703
)

// Список возможных разрешений в системе - один для всех аккаунтов.
var permissions = []Permission{

	// Аккаунт: Роли 101-109

	// Пользователи: 2хх |
	{Name: "Просмотр пользователей", Tag:"user", CodeName: "PermissionUserListing", Code: 201, Description: "Доступ к чтению списка пользователей и их данным"},
	{Name: "Редактирование пользователя", Tag:"user", CodeName: "PermissionUserEditing", Code: 202, Description: "Возможность редактировать роли пользователей"},
	{Name: "Добавление пользователей", Tag:"user", CodeName: "PermissionUserAppend", Code: 203, Description: "Возможность приглашать пользователей в аккаунт."},
	{Name: "Удаление пользователей", Tag:"user", CodeName: "PermissionUserDeleting", Code: 204, Description: "Возможность исключать пользователей из аккаунта."},

	// Склады, товары, услуги: Склады 401-410
	{Name: "Просмотр складов", Tag:"store", CodeName: "PermissionStoreListing", Code: 401, Description: "Доступ к списку складов и их наполнению"},
	{Name: "Редактирование склада", Tag:"store", CodeName: "PermissionStoreEditing", Code: 402, Description: "Возможность редактировать данные созданного склада."},
	{Name: "Создание склада", Tag:"store", CodeName: "PermissionStoreCreating", Code: 403, Description: "Возможность создать склад."},
	{Name: "Удаление склада", Tag:"store", CodeName: "PermissionStoreDeleting", Code: 404, Description: "Возможность удалить склад со всеми его данными."},

	{Name: "Просмотр товаров", Tag:"product", CodeName: "PermissionProductListing", Code: 405, Description: "Доступ к списку товаров во всех складах."},
	{Name: "Редактирование товаров", Tag:"product", CodeName: "PermissionProductEditing", Code: 406, Description: "Редактирование всех данных товаров"},
	{Name: "Создание товаров", Tag:"product", CodeName: "PermissionProductCreating", Code: 407, Description: "Возможность создать новый товар"},
	{Name: "Удаление товаров", Tag:"product", CodeName: "PermissionProductDeleting", Code: 408, Description: "Возможность удалить товар"},

	// Аккаунт 7хх
	{Name: "Управление API-ключами", Tag:"account", CodeName: "PermissionAPIManagement", Code: 700, Description: "Возможность управлять API-ключами (создавать, редактировать и удалять)."},
	{Name: "Управление ролями", Tag:"account", CodeName: "PermissionRoleManagement", Code: 701, Description: "Возможность управлять ролями (создавать, назначать, редактировать и удалять)."},
	{Name: "Управление биллингом", Tag:"account", CodeName: "PermissionBillingManagement", Code: 702, Description: "Возможность управлять платежной информацией"},
	{Name: "Управление аккаунтом", Tag:"account", CodeName: "PermissionAccountManagement", Code: 703, Description: "Администрирование первичных данных аккаунта"},
}

type Permission struct {
	ID        	uint `json:"-" gorm:"primary_key;unique_index;"`
	Name	 	string `json:"name" gorm:"size:255;unique;"` // Просмотр товаров / Редактирование товара / Создание товара / Удаление товара - для удобства
	Tag 		string `json:"tag" gorm:"size:255"` // store, order, leads, contacts...
	CodeName 	string `json:"name" gorm:"size:255;unique_index;"` // PermissionStoreListing, PermissionStoreEditing, ...
	Code 		uint `json:"type" gorm:"unique_index;"`
	Description string `json:"description" gorm:"size:255;"` // Описание права: 'Право на редактирование товара, изображений и т.д.'
	Roles  		[]Role `json:"-" gorm:"many2many:role_permissions;"`
}

func init() {
	// seeding can be: "" / "true" / "fresh"
	seeding := os.Getenv("seeding")
	if seeding == "true" ||  seeding == "fresh"{
		permissionSeeding()
	}
}

// Создание нового правила доступа (сугубо системная функция)
func (permission *Permission) create() error {

	// проверка на попытку создать дубль разрешения, которое уже было создано
	if reflect.TypeOf(permission.ID).String() == "uint" {
		if permission.ID > 0 && !base.GetDB().First(&Permission{}, permission.ID).RecordNotFound() {
			// todo need to translation
			return errors.New("Can't create new role: already crated!")
		}
	}

	// создаем правило
	if err := base.GetDB().Create(permission).Error; err != nil {
		return err
	}
	return nil
}

// Ищет нужный пермишен по заданному коду права
func (permission *Permission) Find(code_int uint) error {
	err := base.GetDB().First(permission, "code = ?", code_int).Error;
	if err != nil || gorm.IsRecordNotFoundError(err) {
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

// заливает первичные права в БД
func permissionSeeding()  {

	if !base.GetDB().Find(&Permission{}).RecordNotFound() {
		return
	}

	for _, v := range permissions {
		err := v.create()
		if err != nil {
			fmt.Println("Cant create Permissions")
		}
	}
}




