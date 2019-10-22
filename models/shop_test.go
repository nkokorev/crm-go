package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistShopTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&Shop{}) {
		tableName := db.NewScope(&Shop{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

