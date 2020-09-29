package models

import (
	"log"
	"time"
)

// обходит email_queue_workflows каждые N секунд
// Считывает по 100 актуальных задач
// Получает данные для отправки писем
// Добавляет письма в MTA-server

func init() {
	// go taskWorker()
}

func taskWorker() {
	
	for {
		if db == nil {
			time.Sleep(time.Millisecond*2000)
			continue
		}

		tasks := make([]TaskScheduler,0)

		// Получаем задачи у которых статус planned
		err := db.Model(&TaskScheduler{}).
			Where("expected_time_to_start <= ? AND status = ?", time.Now().UTC(), WorkStatusPlanned).Limit(10).
			Find(&tasks).Error
		if err != nil {
			log.Printf("TaskScheduler:  %v", err)
			time.Sleep(time.Second*10)
			continue
		}

		// Подготавливаем отправку
		for i := range tasks {
			
			// 1. Перед началом меняем статус у задачи с pending на planned
			if err := tasks[i].SetStatus(WorkStatusPlanned); err != nil {
				log.Printf("taskWorker: Ошибка установки статуса Pending задачи [%v]: %v", tasks[i].Id, err)
				continue
			}

			// 2. Запускаем выполнение задачи и ожидаем результата...
			err = tasks[i].Execute()

			if err != nil {

				// Если задача провалена - ставим статус failed
				if err := tasks[i].SetStatus(WorkStatusFailed, err.Error()); err != nil {
					log.Printf("taskWorker: Ошибка установки статуса Failed у задачи [%v]: %v", tasks[i].Id, err)
				}
				
			}  else {
				// Если задача выполнена - ставим статус completed
				if err := tasks[i].SetStatus(WorkStatusCompleted); err != nil {
					log.Printf("taskWorker: Ошибка установки статуса Completed задачи [%v]: %v", tasks[i].Id, err)
				}
			}

			continue
		}

		if len(tasks) > 100 {
			time.Sleep(time.Minute * 2)
			continue
		}
		
		time.Sleep(time.Second * 20)
		continue
	}
}
