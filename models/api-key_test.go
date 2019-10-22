package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistApiKeyTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&ApiKey{}) {
		tableName := db.NewScope(&ApiKey{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}
