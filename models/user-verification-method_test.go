package models

import (
	"github.com/nkokorev/crm-go/utils"
	"testing"
)

func TestUserVerificationType_Create(t *testing.T) {
	listTest := []struct {
		uvt         UserVerificationMethod
		expected    bool
		description string
	}{
		{UserVerificationMethod{Name: "TestCreate Type", Tag:utils.RandStringBytesMaskImprSrcUnsafe(10, false)}, true, "Все есть"},
		{UserVerificationMethod{Name: "", Tag:utils.RandStringBytesMaskImprSrcUnsafe(9, false)}, false, "Нет имени"},
		{UserVerificationMethod{Name: "TestCreate Type", Tag:"Ф"}, false, "Слишком короткий код"},
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
	uvt, err := UserVerificationMethod{Name: "TestDelete", Tag:utils.RandStringBytesMaskImprSrcUnsafe(5, false)}.Create()
	if err != nil {
		t.Fatalf("Cant create ver %v", err)
	}
	defer uvt.Delete()

	uvtF, err := GetUserVerificationTypeById(uvt.Id)
	if err !=nil {
		t.Fatalf("Cant find ver type by id %v", err)
	}

	if uvt.Tag != uvtF.Tag {
		t.Fatalf("А коды то разные мужик!")
	}
}

func TestGetUserVerificationTypeByCode(t *testing.T) {
	uvt, err := UserVerificationMethod{Name: "TestDelete", Tag:utils.RandStringBytesMaskImprSrcUnsafe(5, false)}.Create()
	if err != nil {
		t.Fatalf("Cant create ver %v", err)
	}
	defer uvt.Delete()

	uvtF, err := GetUserVerificationTypeByCode(uvt.Tag)
	if err != nil {
		t.Fatalf("Cant find ver type by Code %v", err)
	}

	if uvt.Id != uvtF.Id {
		t.Fatalf("А Id то разные мужик!")
	}
}

func TestUserVerificationType_Delete(t *testing.T) {
	uvt, err := UserVerificationMethod{Name: "TestDelete", Tag:"testCode"}.Create()
	if err != nil {
		t.Fatalf("Cant create ver %v", err)
	}
	defer uvt.Delete()

	// убеждаемся, что код верфикации есть
	_, err = GetUserVerificationTypeById(uvt.Id)
	if err !=nil {
		t.Fatalf("Cant find ver type by id %v", err)
	}

	if err := uvt.Delete();err!=nil {
		t.Fatalf("Ай ай, удаление не прошло: %v", err)
	}

	// убеждаемся, что кода верфикации нет
	_, err = GetUserVerificationTypeById(uvt.Id)
	if err ==nil {
		t.Fatalf("Нашли код, хотя он должен был быть удален")
	}

	_, err = GetUserVerificationTypeByCode(uvt.Tag)
	if err ==nil {
		t.Fatalf("Нашли код, хотя он должен был быть удален")
	}
}

