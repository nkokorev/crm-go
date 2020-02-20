package models


type roleType string

const (
	roleTypeGui  roleType = "gui"
	roleTypeApi roleType = "api"
)

type roleAccess string

// Базовая система ролей (9)
const (
	roleOwner  roleAccess = "owner"
	roleAdmin roleAccess = "admin"
	roleManager roleAccess = "manager"
	roleMarketer roleAccess = "marketer"
	roleAuthor roleAccess = "author"
	roleViewer roleAccess = "viewer"
	roleFullAccess roleAccess = "full-access"
	roleSiteAccess roleAccess = "site-access"
	roleReadAccess roleAccess = "read-access"
)

type Role struct {
	ID uint `json:"id" gorm:"primary_key"`
	IssuerAccountId uint `json:"issuerAccountId" gorm:"index;not null;default:1"` // у системных ролей = 1. Из-под RatusCRM аккаунта их можно изменять.
	Tag roleAccess `json:"tag" gorm:"type:varchar(32);not null;"` // client, admin, manager, ...
	Type roleType `json:"type" gorm:"type:varchar(3);not null;"`
	Name string `json:"name" gorm:"type:varchar(255);not null;"` // "Владелец аккаунта", "Администратор", "Менеджер" ...

	Description string `json:"description" gorm:"type:varchar(255);default:null;"` // Краткое описание роли
}

func (Role) PgSqlCreate() {
	db.CreateTable(&Role{})
	db.Exec("ALTER TABLE roles\n    ADD CONSTRAINT users_issuer_account_id_fkey FOREIGN KEY (issuer_account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\ncreate unique index uix_users_issuer_account_id_code ON roles (issuer_account_id,tag);")
}

var roles = []Role {
	{ IssuerAccountId: 1, Name: "Владелец аккаунта",Tag:roleOwner, 		Type:	roleTypeGui,	Description: "Доступ ко всем данным и функционалу аккаунта."},
	{ IssuerAccountId: 1, Name: "Администратор", 	Tag:roleAdmin, 		Type:	roleTypeGui,	Description: "Доступ ко всем данным и функционалу аккаунта. Не может удалить аккаунт или менять владельца аккаунта."},
	{ IssuerAccountId: 1, Name: "Менеджер", 		Tag:roleManager, 	Type:	roleTypeGui,	Description: "Не может добавлять пользователей, менять биллинговую информацию и систему ролей."},
	{ IssuerAccountId: 1, Name: "Маркетолог", 		Tag:roleMarketer, 	Type:	roleTypeGui,	Description: "Читает все клиентские данные, может изменять все что касается маркетинга, но не заказы или склады."},
	{ IssuerAccountId: 1, Name: "Автор", 			Tag:roleAuthor, 	Type:	roleTypeGui,	Description: "Может создавать контент: писать статьи, письма, описания к товарам и т.д."},
	{ IssuerAccountId: 1, Name: "Наблюдатель", 		Tag:roleViewer, 	Type:	roleTypeGui,	Description: "The Viewer can view reports in the account"},
	{ IssuerAccountId: 1, Name: "Full Access", 		Tag:roleFullAccess, Type:	roleTypeApi,	Description: "Доступ ко всем функциям API"},
	{ IssuerAccountId: 1, Name: "Site Access", 		Tag:roleSiteAccess, Type:	roleTypeApi,	Description: "Доступ к аккаунту через API, необходимый для интеграции с сайтом"},
	{ IssuerAccountId: 1, Name: "Read Access", 		Tag:roleReadAccess, Type:	roleTypeApi,	Description: "Доступ к чтению основной информации об аккаунте."},
}