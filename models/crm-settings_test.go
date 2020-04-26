package models

import (
	"testing"
)

func TestGetCrmSettings(t *testing.T) {
	settings, err := GetCrmSettings()
	if err != nil || settings == nil {
		t.Fatalf("Не удалось загрузить настройки crm-системы: %v", err)
	}

	if settings.ID != 1 {
		t.Fatalf("ID у строки настроек не равен 1: , %v", settings)
	}
}

func TestCrmSetting_Save(t *testing.T) {
	settings, err := GetCrmSettings()
	if err != nil {
		t.Fatal("Не удалось получить настройки CRM")
	}

	settings.ApiEnabled = !settings.ApiEnabled
	if err := settings.Save(); err != nil {
		t.Fatalf("Cant Save CRM settings: %v", err)
	}

	// возвращаем назад настройки
	defer func() {
		settings.ApiEnabled = !settings.ApiEnabled
		if err := settings.Save(); err != nil {
			t.Fatalf("Cannot back CRM settings : %v", err)
		}
	}()

	newSettings, err := GetCrmSettings()
	if err != nil {
		t.Fatal("Не удалось получить настройки CRM (2)")
	}

	if newSettings.ApiEnabled != settings.ApiEnabled {
		t.Fatal("Функция Save CrmSettings реально не сохранила данные")
	}
}

func BenchmarkGetCrmSettings(b *testing.B) {
	settings, err := GetCrmSettings()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		settings, err = GetCrmSettings()
		if err != nil || settings == nil {
			b.Fatalf("Не удалось загрузить настройки crm-системы: %v", err)
		}
	}
}
