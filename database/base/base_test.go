package base

import (
	"github.com/nkokorev/crm-go/models"
	"reflect"
	"testing"
)

func TestGetDB(t *testing.T) {
	db := models.GetDB()
	if reflect.TypeOf(db).String() != "*gorm.DB" {
		t.Error("expected type: *gorm.DB, get: ", reflect.TypeOf(db).String())
	}
}

