package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistAccountTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&Account{}) {
		tableName := db.NewScope(&Account{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

