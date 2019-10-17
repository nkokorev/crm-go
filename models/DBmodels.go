package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/database"
	u "github.com/nkokorev/crm-go/utils"
	"log"
	"reflect"
	"time"
)

// public model
type DBModel struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

// создает и возвращает уникальный новый хеш для модели IFace interface{}
func CreateHashID(IFace interface{}) (hash string, error u.Error) {

	model := reflect.ValueOf(IFace)

	if model.Type().Kind() != reflect.Ptr {
		fmt.Println("Model has not HashID!")
		error.Message = "Model has not HashID!"
		return
	}

	Field := model.Elem().FieldByName("HashID")
	if !Field.IsValid() {
		fmt.Println("Model has not HashID!")
		error.Message = "Model has not HashID!"
		return
	}

	db := database.GetDB()

	hash = u.RandStringBytes(u.LENGTH_HASH_ID); count := 0
	for i:= 0; i < 5; i++ {
		hash = u.RandStringBytes(u.LENGTH_HASH_ID)
		db.Model(IFace).Where("hash_id = ?", hash).Count(&count)
		if count == 0 {
			break
		} else {
			log.Println("Неудалось создать новый hash_id", hash)
		}
	}
	return
}