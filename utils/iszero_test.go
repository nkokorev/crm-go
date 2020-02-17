package utils

import (
	"testing"
)

type TestStruct struct {
	Username string
	Email string
	Phone string
	Name string
}

func TestAccount_CheckUserInputRequiredFields(t *testing.T) {

	TestList := []struct {
		Enum []string
		Input interface{}
		expected bool
		description string
	}{
		{
			[]string{"username","email","phone"},
			&TestStruct{Username:"", Email:"", Phone:""},
			false,
			"Требуемые поля - пустые",
		},
		{
			[]string{"email","username"},
			&TestStruct{Username:"TestUser 1", Email:"mail@example.com"},
			true,
			"Формально поля есть",
		},
		{
			[]string{"name","phone"},
			&TestStruct{Email:"kokorevn@gmail.com", Name:"Никита"},
			false,
			"Нет поля с телефоном",
		},
		{
			[]string{"name","phone","username"},
			&TestStruct{Username:"",Phone:"+79251952295", Name:"Никита"},
			false,
			"Нет поля с телефоном",
		},
	}

	for i, v := range TestList {
		err := CheckNotNullFields(v.Input, v.Enum)

		// если прошел проверку
		if v.expected == true && err != nil {
			t.Fatalf("Проверка провалена, а должна была пройти:\nПользователь %v : \nОжидалось: %v \n user: %v \nТребуемые поля: %v", i, v.expected, v.Input, v.Enum)
		}

		if v.expected == false && err == nil {
			t.Fatalf("Проверка прошла успешно, но должна быть провалена:\nПользователь %v : \nОжидалось: %v \n user: %v \nТребуемые поля: %v", i, v.expected, v.Input, v.Enum)
		}

	}

}
