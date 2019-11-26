package models

import "testing"

// ### Testing CRUD functions

func TestUser_Create(t *testing.T) {

	user := User{ Username:"TestuserName", Email:"testmail@ratus-dev.ru", Password:"qwert123_QWR", Name:"Test User", }

	if !DB.NewRecord(user) || !DB.NewRecord(&user) {
		t.Error("User should be new record before create")
	}
}

func TestUser_Save(t *testing.T) {

}

func TestUser_Update(t *testing.T) {

}

func TestUser_Get(t *testing.T) {

}

func TestUser_Delete(t *testing.T) {

}
