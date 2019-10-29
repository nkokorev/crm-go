package rebase

import (
	"github.com/nkokorev/crm-go/database/migrations"
	"github.com/nkokorev/crm-go/database/seeds"
	"github.com/nkokorev/crm-go/models"
)

func Rebase()  {
	migrations.MigrationTables(true)
	models.PermissionSeeding()
	models.RoleSeeding()
	seeds.UserSeeding()
}
