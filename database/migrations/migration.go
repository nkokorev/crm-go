package migrations

import (
	"github.com/nkokorev/crm-go/database/base"
	"github.com/nkokorev/crm-go/models"
)

func MigrationTables(freshTables bool) {

	db := base.GetDB()

	// ломать не строить (последовательность важна)
	if freshTables {

		db.DropTableIfExists("account_user_roles")
		db.DropTableIfExists("role_permissions")
		db.DropTableIfExists("account_users")
		db.DropTableIfExists("user_roles")
		db.DropTableIfExists(&models.Product{}, &models.Shop{},&models.ApiKey{}, &models.Role{}, &models.Permission{}, &models.Role{}, &models.Store{}, &models.AccountUser{}, &models.Account{}, &models.User{}, )
	}

	// теперь создаем таблички
	db.Debug().AutoMigrate(&models.AccountUser{}, &models.User{}, &models.Account{}, &models.Store{}, &models.Permission{}, &models.Role{}, &models.ApiKey{}, &models.Shop{},&models.Product{})

	db.Table("accounts").AddForeignKey("user_id", "users(id)", "RESTRICT", "CASCADE") // за пользователем удаляются все его аккаунты

	//db.Table("users").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	db.Table("account_users").AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	db.Table("account_users").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Table("account_users").AddForeignKey("role_id", "roles(id)", "RESTRICT", "CASCADE") // нельзя удалить роль, если к ней привязан хотя бы 1 пользователь

	db.Table("roles").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Table("role_permissions").AddForeignKey("role_id", "roles(id)", "CASCADE", "CASCADE")
	db.Table("role_permissions").AddForeignKey("permission_id", "permissions(id)", "CASCADE", "CASCADE")

	db.Table("api_keys").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Table("api_keys").AddForeignKey("role_id", "roles(id)", "RESTRICT", "CASCADE") // нельзя удалить роль, если к ней привязан хотя бы 1 ключ

	//

	db.Table("products").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
