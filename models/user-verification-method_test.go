package models

import (
	"github.com/nkokorev/crm-go/utils"
	"testing"
)

func TestUserVerificationType_Create(t *testing.T) {
	listTest := []struct {
		uvt UserVerificationMethod
		expected bool
		description string
	}{
		{UserVerificationMethod{Name:"TestCreate Type", Code:utils.RandStringBytesMaskImprSrcUnsafe(10, false)}, true, "Все есть"},
		{UserVerificationMethod{Name:"", Code:utils.RandStringBytesMaskImprSrcUnsafe(9, false)}, false, "Нет имени"},
		{UserVerificationMethod{Name:"TestCreate Type", Code:"Ф"}, false, "Слишком короткий код"},
	}

	for i,v := range listTest {
		uvt,err := v.uvt.Create()

		if v.expected == true && err != nil {
			t.Fatalf("Не удалось создать настройки верификации для варианта %v, ошибка: %v", i, err)
		}

		if v.expected == false && err == nil {
			t.Fatalf("Удалось создать настройки верификации, хотя они не должны были создаться для варианта %v", i)

		}

		if uvt != nil && err == nil {
			defer uvt.Delete()
		}

	}
}

func TestGetUserVerificationTypeById(t *testing.T) {
	uvt, err := UserVerificationMethod{Name:"TestDelete", Code:utils.RandStringBytesMaskImprSrcUnsafe(5, false)}.Create()
	if err != nil {
		t.Fatalf("Cant create ver %v", err)
	}
	defer uvt.Delete()

	uvtF, err := GetUserVerificationTypeById(uvt.ID)
	if err !=nil {
		t.Fatalf("Cant find ver type by id %v", err)
	}

	if uvt.Code != uvtF.Code {
		t.Fatalf("А коды то разные мужик!")
	}
}

func TestGetUserVerificationTypeByCode(t *testing.T) {
	uvt, err := UserVerificationMethod{Name:"TestDelete", Code:utils.RandStringBytesMaskImprSrcUnsafe(5, false)}.Create()
	if err != nil {
		t.Fatalf("Cant create ver %v", err)
	}
	defer uvt.Delete()

	uvtF, err := GetUserVerificationTypeByCode(uvt.Code)
	if err !=nil {
		t.Fatalf("Cant find ver type by Code %v", err)
	}

	if uvt.ID != uvtF.ID {
		t.Fatalf("А ID то разные мужик!")
	}
}

func TestUserVerificationType_Delete(t *testing.T) {
	uvt, err := UserVerificationMethod{Name:"TestDelete", Code:"testCode"}.Create()
	if err != nil {
		t.Fatalf("Cant create ver %v", err)
	}
	defer uvt.Delete()

	// убеждаемся, что код верфикации есть
	_, err = GetUserVerificationTypeById(uvt.ID)
	if err !=nil {
		t.Fatalf("Cant find ver type by id %v", err)
	}

	if err := uvt.Delete();err!=nil {
		t.Fatalf("Ай ай, удаление не прошло: %v", err)
	}

	// убеждаемся, что кода верфикации нет
	_, err = GetUserVerificationTypeById(uvt.ID)
	if err ==nil {
		t.Fatalf("Нашли код, хотя он должен был быть удален")
	}

	_, err = GetUserVerificationTypeByCode(uvt.Code)
	if err ==nil {
		t.Fatalf("Нашли код, хотя он должен был быть удален")
	}
}


