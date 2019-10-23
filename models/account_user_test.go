package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistAccountUserTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&AccountUser{}) {
		tableName := db.NewScope(&AccountUser{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}
