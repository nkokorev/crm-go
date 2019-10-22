package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistStoreTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&Store{}) {
		tableName := db.NewScope(&Store{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

