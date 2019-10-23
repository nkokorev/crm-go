package utils

import (
	"errors"
	"github.com/nkokorev/crm-go/database/base"
	"math/rand"
	"reflect"
	"time"
)

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyz"
const LENGTH_HASH_ID  = 8


func RandStringBytes(n int) string {
	return StringWithCharset(n, letterBytes)
}

// Создает уникальный hashId для модели. Получать указатель на модель
func CreateHashID(IFace interface{}) (hash string, error error) {

	model := reflect.ValueOf(IFace)

	if model.Type().Kind() != reflect.Ptr {
		return "", errors.New("Model is not Ptr (*)")
	}

	Field := model.Elem().FieldByName("HashID")
	if !Field.IsValid() {
		return "", errors.New("Model has not fiels: HashID")
	}

	db := base.GetDB()

	count := 0
	for i:= 0; i < 5; i++ {
		hash = RandStringBytes(LENGTH_HASH_ID)
		db.Model(IFace).Where("hash_id = ?", hash).Count(&count)
		if count == 0 {
			break
		}
		hash = RandStringBytes(LENGTH_HASH_ID)
	}
	if count != 0 {
		return "", errors.New("Cant create new hash for Model")
	}
	return
}

func StringWithCharset(length int, charset string) string {

	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}