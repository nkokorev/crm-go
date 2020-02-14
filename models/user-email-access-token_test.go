package models

import (
	"testing"
	"time"
)

func TestEmailAccessToken_Expired(t *testing.T) {

	// todo дописать список тестов
	var testList = []struct{
		eat EmailAccessToken
		expected bool
		description string
	}{
		{EmailAccessToken{CreatedAt:time.Now()}, false, "Т.к. время текущее, то токен точно не должен быть просрочен"},
		{EmailAccessToken{CreatedAt:time.Now().Truncate(time.Hour*24*100)}, true, "Токен должен быть просрочен т.к. ему уже 100 дней"},
	}

	for _,v := range testList {

		if v.eat.Expired() != v.expected {
			t.Fatalf("Не верная работа функции для: %v \n Описание: %v", v, v.description)
		}
	}
}
