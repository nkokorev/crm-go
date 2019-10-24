package migrations

import (
	"github.com/nkokorev/crm-go/database/base"
	"github.com/nkokorev/crm-go/models"
	"os"
)

func init() {

	// migration can be: "" / "true" / "fresh"
	migration := os.Getenv("migration")
	if migration == "true" || migration == "fresh" {
		//MigrationTables(migration == "fresh")
		MigrationTables(true)
	}
	//MigrationTables(true)
}


func MigrationTables(freshTables bool) {

	db := base.GetDB()

	// ломать не строить (последовательность важна)
	if freshTables {
		db.DropTableIfExists("api_key_permissions")
		db.DropTableIfExists("account_user_roles")
		db.DropTableIfExists("role_permissions")
		db.DropTableIfExists("account_user_permissions")
		db.DropTableIfExists("account_users")
		db.DropTableIfExists("user_roles")
		db.DropTableIfExists(&models.Shop{},&models.ApiKey{}, &models.Role{}, &models.Permission{}, &models.Role{}, &models.Store{}, &models.AccountUser{}, &models.Account{}, &models.User{})
	}

	// теперь создаем таблички
	db.Debug().AutoMigrate(&models.AccountUser{}, &models.User{}, &models.Account{}, &models.Store{}, &models.Permission{}, &models.Role{}, &models.ApiKey{}, &models.Shop{})

	db.Table("accounts").AddForeignKey("user_id", "users(id)", "RESTRICT", "CASCADE") // за пользователем удаляются все его аккаунты

	db.Table("account_users").AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	db.Table("account_users").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	//db.Table("account_user_permissions").AddForeignKey("account_user_id", "account_users(id)", "CASCADE", "CASCADE")
	//db.Table("account_user_permissions").AddForeignKey("permission_id", "permissions(id)", "CASCADE", "CASCADE")

	db.Table("account_user_roles").AddForeignKey("account_user_id", "account_users(id)", "CASCADE", "CASCADE")
	db.Table("account_user_roles").AddForeignKey("role_id", "roles(id)", "CASCADE", "CASCADE")

	db.Table("roles").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	db.Table("role_permissions").AddForeignKey("role_id", "roles(id)", "CASCADE", "CASCADE")
	db.Table("role_permissions").AddForeignKey("permission_id", "permissions(id)", "CASCADE", "CASCADE")

	db.Table("api_key_permissions").AddForeignKey("api_key_id", "api_keys(id)", "CASCADE", "CASCADE")
	db.Table("api_key_permissions").AddForeignKey("permission_id", "permissions(id)", "CASCADE", "CASCADE")
}