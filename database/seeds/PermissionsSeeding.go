package seeds

import (
	"fmt"
	"github.com/nkokorev/crm-go/database/base"
	"github.com/nkokorev/crm-go/models"
)

var permissions = []models.Permission{

	// Аккаунт: Роли 101-109

	// Пользователи: 2хх |
	{Name: "Просмотр пользователей", Tag:"user", CodeName: "PermissionUserListing", Code: 201, Description: "Доступ к чтению списка пользователей и их данным"},
	{Name: "Редактирование пользователя", Tag:"user", CodeName: "PermissionUserEditing", Code: 202, Description: "Возможность редактировать права пользователя"},
	{Name: "Добавление пользователей", Tag:"user", CodeName: "PermissionUserAppend", Code: 203, Description: "Возможность приглашать пользователей в аккаунт."},
	{Name: "Удаление пользователей", Tag:"user", CodeName: "PermissionUserDeleting", Code: 204, Description: "Возможность исключать пользователей из аккаунта."},

	// Склады, товары, услуги: Склады 401-410
	{Name: "Просмотр складов", Tag:"store", CodeName: "PermissionStoreListing", Code: 401, Description: "Доступ к списку складов"},
	{Name: "Редактирование склада", Tag:"store", CodeName: "PermissionStoreEditing", Code: 402, Description: "Возможность редактировать данные уже созданного склада."},
	{Name: "Создание склада", Tag:"store", CodeName: "PermissionStoreCreating", Code: 403, Description: "Возможность создать склад."},
	{Name: "Удаление склада", Tag:"store", CodeName: "PermissionStoreDeleting", Code: 404, Description: "Возможность удалить склад со всеми его данными."},

	{Name: "Просмотр товаров", Tag:"product", CodeName: "PermissionProductListing", Code: 401, Description: "Доступ к списку товаров"},
	{Name: "Редактирование товаров", Tag:"product", CodeName: "PermissionProductEditing", Code: 402, Description: "Редактирование товаров"},
	{Name: "Создание товаров", Tag:"product", CodeName: "PermissionProductCreating", Code: 403, Description: "Возможность создать новый товар"},
	{Name: "Удаление товаров", Tag:"product", CodeName: "PermissionProductDeleting", Code: 404, Description: "Возможность удалить товар"},

	// Аккаунт 7хх
	{Name: "Управление ролями", Tag:"account", CodeName: "PermissionRoleManagement", Code: 701, Description: "Возможность управлять ролями (создавать, редактировать и удалять)."},
	{Name: "Управление API-ключами", Tag:"account", CodeName: "PermissionAPIManagement", Code: 702, Description: "Возможность управлять API-ключами (создавать, редактировать и удалять)."},
	{Name: "Администратор аккаунта", Tag:"account", CodeName: "PermissionAccountManagement", Code: 777, Description: "Администратор аккаунта, который имеет полный доступ к системе."},
}

// разворачивает базовые разрешения для всех пользователей
func PermissionSeeding()  {

	db := base.GetDB()

	db.Unscoped().Delete(models.Permission{})

	for _, v := range permissions {
		err := v.Create()
		if err != nil {
			fmt.Println("Cant create Permissions")
		}
	}
}
