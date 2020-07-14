package models

import "reflect"

// Объекты для хранения в БД списка событий и
// Возвращает список всех возможных обработчиков
func GetSystemHandlerList () []map[string]string{
	return []map[string]string {
		{"name": "EmailQueueRun", "description": "Запуск автоматической серии email писем с указанным ID."},
		{"name":"WebHookCall","description":"Вызов WebHook с указанным ID."},
	}
}

// Возвращает список доступных событий в системе
func GetSystemEventList() []string { // el := EventList{}
	list := make([]string,0)

	t := reflect.TypeOf(&Event{})
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		list = append(list, m.Name)
	}

	return list
}