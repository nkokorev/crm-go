package rebase

import (
	"os"
)

func Rebase()  {

	// на всякий случай выход, если это не тест
	if os.Getenv("ENV_VAR") != "test" {
		return
	}

	//migrations.MigrationTables(true)
	//migrations.MigrationTables2(true)
	//migrations.MigrationTables3(true)


	//models.PermissionSeeding()
	//models.CreateSystemRoles()
	//seeds.UserSeeding()
	//models.CreateSystemEavAttr()
}
