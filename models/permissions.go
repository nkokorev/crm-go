package models

type Permission struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	AccountID        uint `gorm:"index;" json:"-"`
	OwnerID uint
	OwnerType string // User, UserRole, ApiToken
	//HashID string `json:"hash_id" gorm:"type:varchar(10);unique_index;"`
	Code uint `json:"code"`
	Name string `json:"name" gorm:"size:255"` // Просмотр товаров / Редактирование товара / Создание товара / Удаление товара
	Status bool `json:"status"` // allow - true, deny - false (разрешить / запретить)
}

const (
	PermissionStoreListing           = 101 // Доступ к списку складов
	PermissionStoreEditing           = 102 // Редактирование данных склада
	PermissionStoreCreating          = 103 // Создание склада
	PermissionStoreDeleting          = 104 // Удаление склада
)

// устанавливает все базовые значения разрешений, в соответствии с ролью (Admin, ClientsManager, StoreKeeper)
func (user *User) PermissionInitial(account *Account, role string) {

	return
}

func (user *User) PermissionCheck(CheckedPermissions uint) (status bool) {
	status = false // default
	return
}



type UserGroup struct {
	ID string `json:"hash_id" gorm:"type:varchar(10);primary_key;unique_index;"`
	Name string `json:"name" gorm:"size:255"` // Администраторы / Менеджеры по работе с клиентами / Управление складом / Бухгалтерия
}

/**
* Ролевая модель определяет объединяет в группу права
* Доступ через роль: users | user <> roles | role <> permissions
* При пересечении ролей, определяющей будет роль с наиболее широкими правами (модель perm.Status || perm-2.Status || ...)
* Однако, при определении прав доступа будет приоритетным прямое назначение: users | user <> permissions
 */
type Role struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID string `json:"hash_id" gorm:"type:varchar(10);unique_index;"`

	AccountID uint `json:"account_id" gorm:"index;"` // к какому аккаунту принадлежит (пользователь может быть в 2-х системах одновременно)
	Name string `json:"name" gorm:"size:255"` // Администратор / Менеджер / Оператор / Кладовщик / Маркетолог
	Description string `json:"name" gorm:"size:255;"` // 'Роль для новых администраторов склада, без возможности удаления или создания новых складов...'
	//UserID uint `json:"user_id" gorm:"index;"`
	Users []User `json:"user_id" gorm:"many2many:user_roles;"`
	Permissions []Permission `json:"permissions"` // Has many []
}
