package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistPermissionTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&Permission{}) {
		tableName := db.NewScope(&Permission{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

func TestPermission_Find(t *testing.T) {
	// todo дописать код проверки функции
}

