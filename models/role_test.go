package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistRoleTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&Role{}) {
		tableName := db.NewScope(&Role{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

