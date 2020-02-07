package models

import "log"

// подключаемся к БД
func init() {

	if Connect() == nil {
		log.Fatal("DB is not loaded")
	}
}

