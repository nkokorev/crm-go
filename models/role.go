package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
)

type roleType string

const (
	roleTypeGui roleType = "gui"
	roleTypeApi roleType = "api"
)

type AccessRole = string

// Базовая система ролей (9)
const (
	RoleOwner      AccessRole = "owner"
	RoleAdmin      AccessRole = "admin"
	RoleManager    AccessRole = "manager"
	RoleMarketer   AccessRole = "marketer"
	RoleAuthor     AccessRole = "author"
	RoleViewer     AccessRole = "viewer"
	RoleClient     AccessRole = "client"
	RoleFullAccess AccessRole = "full-access"
	RoleSiteAccess AccessRole = "site-access"
	RoleReadAccess AccessRole = "read-access"
)

type Role struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"type:int;index;not null;"` // у системных ролей = 1. Из-под RatusCRM аккаунта их можно изменять.

	// IssuerAccountId uint       `json:"issuerAccountId" gorm:"index;not null;default:1"`
	Tag             AccessRole `json:"tag" gorm:"type:varchar(32);not null;"`	// client, admin, manager, ...
	Type            roleType   `json:"type" gorm:"type:varchar(3);not null;"`	// gui / api
	Name            string     `json:"name" gorm:"type:varchar(255);not null;"` // "Владелец аккаунта", "Администратор", "Менеджер" ...

	Description string `json:"description" gorm:"type:varchar(255);default:null;"` // Краткое описание роли
}

var systemRoles = []Role{
	{ AccountID: 1, Name: "Владелец аккаунта",Tag: RoleOwner, 	Type: roleTypeGui,	Description: "Доступ ко всем данным и функционалу аккаунта."},
	{ AccountID: 1, Name: "Администратор", 	Tag: RoleAdmin, 	Type: roleTypeGui,	Description: "Доступ ко всем данным и функционалу аккаунта. Не может удалить аккаунт или менять владельца аккаунта."},
	{ AccountID: 1, Name: "Менеджер", 		Tag: RoleManager, 	Type: roleTypeGui,	Description: "Не может добавлять пользователей, менять биллинговую информацию и систему ролей."},
	{ AccountID: 1, Name: "Маркетолог", 		Tag: RoleMarketer, 	Type: roleTypeGui,	Description: "Читает все клиентские данные, может изменять все что касается маркетинга, но не заказы или склады."},
	{ AccountID: 1, Name: "Автор", 			Tag: RoleAuthor, 	Type: roleTypeGui,	Description: "Может создавать контент: писать статьи, письма, описания к товарам и т.д."},
	{ AccountID: 1, Name: "Наблюдатель", 		Tag: RoleViewer, 	Type: roleTypeGui,	Description: "The Viewer can view reports in the account"},
	{ AccountID: 1, Name: "Клиент", 			Tag: RoleClient, 	Type: roleTypeGui,	Description: "Стандартная роль для всех клиентов"},
	{ AccountID: 1, Name: "Full Access", 		Tag: RoleFullAccess, Type: roleTypeApi,	Description: "Доступ ко всем функциям API"},
	{ AccountID: 1, Name: "Site Access", 		Tag: RoleSiteAccess, Type: roleTypeApi,	Description: "Доступ к аккаунту через API, необходимый для интеграции с сайтом"},
	{ AccountID: 1, Name: "Read Access", 		Tag: RoleReadAccess, Type: roleTypeApi,	Description: "Доступ к чтению основной информации об аккаунте."},
}

func (Role) PgSqlCreate() {
	db.CreateTable(&Role{})
	db.Exec("-- ALTER TABLE roles\n--     ADD CONSTRAINT roles_issuer_account_id_fkey FOREIGN KEY (issuer_account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\ncreate unique index uix_roles_issuer_account_id_tag_code ON roles (account_id, tag);")
	db.Model(&Role{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	// Создаем системные роли
	for i, v := range systemRoles {
		_, err := v.create();
		if err != nil {
			log.Fatalf("Cant create role[%v]: %v", i, err)
		}
	}
}
func (role *Role) BeforeCreate(scope *gorm.Scope) error {
	role.ID = 0
	return nil
}


// ############# Entity interface #############
func (role Role) GetId() uint { return role.ID }
func (role *Role) setId(id uint) { role.ID = id }
func (role Role) GetAccountId() uint { return role.AccountID }
func (role *Role) setAccountId(id uint) { role.AccountID = id }
func (role Role) systemEntity() bool {
	return role.AccountID == 1
}
// ############# Entity interface #############

func (role Role) create() (Entity, error)  {

	// Validate
	if role.AccountID < 1 {
		return nil, utils.Error{Message:"Не корректно указаны данные", Errors: map[string]interface{}{"roleIssuerAccountId":"Необходимо указать выпускающий аккаунт!"}}
	}
	if len([]rune(role.Tag)) > 32 || len([]rune(role.Tag)) < 3 {
		return nil, utils.Error{Message:"Не корректно указаны данные", Errors: map[string]interface{}{"roleTag":"Тип должен быть от 3 до 32 символов!"}}
	}
	if role.Type != roleTypeGui && role.Type != roleTypeApi {
		return nil, utils.Error{Message:"Не корректно указаны данные", Errors: map[string]interface{}{"roleType":"Тип должен быть или gui или api!"}}
	}

	if len([]rune(role.Name)) < 3 {
		return nil, utils.Error{Message:"Не корректно указаны данные", Errors: map[string]interface{}{"roleName":"Слишком короткое имя, минимум 3 символа"}}
	}
	if len([]rune(role.Name)) > 255 {
		return nil, utils.Error{Message:"Не корректно указаны данные", Errors: map[string]interface{}{"roleName":"Слишком длинное имя, макс. 255 символов"}}
	}
	if len([]rune(role.Description)) > 255 {
		return nil, utils.Error{Message:"Не корректно указаны данные", Errors: map[string]interface{}{"roleDescription":"Слишком длинное описание, макс. 255 символов"}}
	}

	var newItem Entity = &role

	if err := db.Create(newItem).First(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (Role) get(id uint) (Entity, error) {

	var role Role

	err := db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (role *Role) load() error {

	err := db.First(role).Error
	if err != nil {
		return err
	}
	return nil
}

func (Role) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	roles := make([]Role,0)
	var total uint

	err := db.Model(&Role{}).Order(sortBy).Limit(1000).
		Where("account_id IN (?)", []uint{1, accountId}).
		Find(&roles).Error
	if err != nil {
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&Role{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(roles))
	for i, v := range roles {
		entities[i] = &v
	}

	return entities, total, nil
}
func (Role) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	roles := make([]Role,0)
	var total uint

	if len(search) > 0 {

		search = "%"+search+"%"

		err := db.Model(&Role{}).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&roles, "tag ILIKE ? OR type ILIKE ? OR name ILIKE ? OR description ILIKE ?", search,search,search,search).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Role{}).
			Where("account_id IN (?) AND tag ILIKE ? OR type ILIKE ? OR name ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}


	} else {
		err := db.Model(&Role{}).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&roles).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Role{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(roles))
	for i, v := range roles {
		entities[i] = &v
	}

	return entities, total, nil
}

func (role *Role) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(role).Omit("id", "account_id").Updates(input).Error
}

func (role Role) delete () error {
	return db.Model(Role{}).Where("id = ?", role.ID).Delete(role).Error
}

////////////

func (role Role) IsOwner () bool {
	return role.Tag == "owner" && role.AccountID == 1
}
func (role Role) IsAdmin () bool {
	return role.Tag == "admin" && role.AccountID == 1
}

// Получаем роль либо системную или в акаунте. Как правило, первое.
func (account Account) GetRole (roleId uint) (*Role, error) {
	var role Role
	err := db.Where("account_id IN (?) AND id = ?", []uint{1, account.ID}, roleId).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (account Account) GetRoleByTag (tag AccessRole) (*Role, error) {
	var role Role
	err := db.Where("account_id IN (?) AND tag = ?", []uint{1, account.ID}, tag).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}



// есть контроллер
func (account Account) GetRoleList() ([]Role, error) {
	roles := make([]Role, 0)

	err := db.Find(&roles, "account_id IN (?)", []uint{1, account.ID}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return roles, nil
}

// хз
func (role Role) Exist() bool {
	return !db.Model(&Role{}).Unscoped().First(&Role{}, role.ID).RecordNotFound()
}

// хз хз хз 
func GetRole(tag AccessRole) (*Role, error) {
	var role Role
	if err := db.Model(&Role{}).First(&role, "account_id = ? AND tag = ?", 1, tag).Error; err != nil {
		return nil, err
	}

	 return &role, nil
}

