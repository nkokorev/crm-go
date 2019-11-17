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
		db.DropTableIfExists(
			&models.ApiKey{}, &models.Role{}, &models.Permission{}, &models.Role{}, &models.Store{}, &models.AccountUser{}, &models.Account{}, &models.User{}, )
	}

	// теперь создаем таблички
	db.Debug().AutoMigrate(&models.AccountUser{}, &models.User{}, &models.Account{}, &models.Store{}, &models.Permission{}, &models.Role{}, &models.ApiKey{})

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



}

func MigrationTables2(freshTables bool)  {
	db := base.GetDB()

	if freshTables {

		db.DropTableIfExists("category_products")
		db.DropTableIfExists(&models.Product{},&models.Category{}, &models.Store{}, &models.StoreWebSite{})
	}

	db.Debug().AutoMigrate( &models.Store{}, &models.StoreWebSite{}, &models.Product{}, &models.Category{}, )

	db.Table("products").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Table("products").AddForeignKey("category_id", "categories(id)", "CASCADE", "CASCADE")
	db.Table("products").AddForeignKey("product_price_id", "product_prices(id)", "CASCADE", "CASCADE")

	db.Table("categories").AddForeignKey("parent_id", "categories(hash_id)", "CASCADE", "CASCADE")
	db.Table("categories").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Table("categories").AddForeignKey("store_id", "stores(hash_id)", "CASCADE", "CASCADE")
}


func MigrationTables3(freshTables bool)  {

	db := base.GetDB()

	if freshTables {

		db.DropTableIfExists("category_products")


		db.DropTableIfExists("eav_attr_set_products")
		db.DropTableIfExists("eav_attr_products")
		db.DropTableIfExists("eav_attr_input_types")
		db.DropTableIfExists("eav_attr_types")
		db.DropTableIfExists("eav_attr_eav_attr_sets")
		db.DropTableIfExists("eav_attr_varchar")
		db.DropTableIfExists("eav_attrs")

		db.DropTableIfExists(&models.EavAttrSet{}, &models.EavAttrInputType{}, &models.EavAttrType{}, &models.EavAttrVarchar{}, &models.EavAttr{},
			&models.Category{}, &models.Product{}, models.Store{}, &models.StoreWebSite{},   )

	}

	db.AutoMigrate(&models.Product{}, &models.Category{}, &models.Store{}, &models.StoreWebSite{},
	&models.EavAttr{}, &models.EavAttrVarchar{}, &models.EavAttrType{}, &models.EavAttrInputType{}, &models.EavAttrSet{})



	db.Table("products").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Table("products").AddForeignKey("category_id", "categories(id)", "CASCADE", "CASCADE")
	db.Table("products").AddForeignKey("product_price_id", "product_prices(id)", "CASCADE", "CASCADE")
	db.Table("products").AddForeignKey("eav_attr_set_id", "eav_attr_sets(id)", "CASCADE", "CASCADE")

	db.Table("categories").AddForeignKey("parent_id", "categories(hash_id)", "CASCADE", "CASCADE")
	db.Table("categories").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Table("categories").AddForeignKey("store_id", "stores(hash_id)", "CASCADE", "CASCADE")


	db.Table("eav_attrs").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Table("eav_attrs").AddForeignKey("eav_attr_type_id", "eav_attr_types(id)", "CASCADE", "CASCADE")
	db.Table("eav_attrs").AddForeignKey("eav_attr_input_type_id", "eav_attr_input_types(id)", "CASCADE", "CASCADE")


	db.Table("eav_attr_varchar").AddForeignKey("eav_attr_id", "eav_attrs(id)", "CASCADE", "CASCADE")


	//db.Table("eav_attr_eav_attr_groups").AddForeignKey("eav_attr_id", "eav_attrs(id)", "CASCADE", "CASCADE")
	//db.Table("eav_attr_eav_attr_groups").AddForeignKey("eav_attr_group_id", "eav_attr_groups(id)", "CASCADE", "CASCADE")

	db.Table("eav_attr_eav_attr_sets").AddForeignKey("eav_attr_id", "eav_attrs(id)", "CASCADE", "CASCADE")
	db.Table("eav_attr_eav_attr_sets").AddForeignKey("eav_attr_set_id", "eav_attr_sets(id)", "CASCADE", "CASCADE")

	//db.Table("eav_attr_group_products").AddForeignKey("product_id", "products(id)", "CASCADE", "CASCADE")
	//db.Table("eav_attr_group_products").AddForeignKey("eav_attr_group_id", "eav_attr_groups(id)", "CASCADE", "CASCADE")
	db.Table("eav_attr_products").AddForeignKey("eav_attr_id", "eav_attrs(id)", "CASCADE", "CASCADE")
	db.Table("eav_attr_products").AddForeignKey("product_id", "products(id)", "CASCADE", "CASCADE")

	db.Table("eav_attr_set_products").AddForeignKey("eav_attr_id", "eav_attrs(id)", "CASCADE", "CASCADE")
	db.Table("eav_attr_set_products").AddForeignKey("product_id", "products(id)", "CASCADE", "CASCADE")

	//db.Table("eav_attr_set").AddForeignKey("product_id", "products(id)", "CASCADE", "CASCADE")



}
