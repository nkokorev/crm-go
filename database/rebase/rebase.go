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
	//models.PermissionSeeding()
	//models.RoleSeeding()
	//seeds.UserSeeding()
}
