package models

import (
	"fmt"
	"testing"
)

func TestCrmSetting_Create(t *testing.T) {

	existSettings := !db.Model(&CrmSetting{}).Find(&CrmSetting{}, "id = 1").RecordNotFound()

	_, err := CreateCrmSettings()

	// 1. Вариант 1-й, не должны создаться настройки
	if existSettings && err == nil {
		t.Error("Настройки crm системы создались, хоть в системе они уже есть")
	}

	if !existSettings && err != nil {
		t.Error("Неудалось создать настройки CRM-системы")
	}

}

func BenchmarkHello(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("hello")
	}
}
