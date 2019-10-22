package base

import (
	"reflect"
	"testing"
)

func TestGetDB(t *testing.T) {
	db := GetDB()
	if reflect.TypeOf(db).String() != "*gorm.DB" {
		t.Error("expected type: *gorm.DB, get: ", reflect.TypeOf(db).String())
	}
}

