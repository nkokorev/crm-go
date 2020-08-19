package models

import (
	"log"
	"time"
)

// обходит email_queue_workflows каждые N секунд
// Считывает по 100 актуальных задач
// Получает данные для отправки писем
// Добавляет письма в MTA-server

var withoutMistakesTaskWorker = true

func init() {
	go taskWorker()
}

func taskWorker() {
	
	for {
		if db == nil {
			time.Sleep(time.Millisecond*2000)
			continue
		}
		if !withoutMistakesTaskWorker {
			log.Println("Какие-то ошибки выполнения taskWorker")
			time.Sleep(time.Minute*5)
			continue
		}

		// fmt.Println("Обход taskWorker")
		
		tasks := make([]TaskScheduler,0)

		// Получаем задачи у которых серия запущена
		err := db.Model(&TaskScheduler{}).
			Where("expected_time_to_start <= ? AND status = 'planned'", time.Now().UTC()).Limit(10).
			Find(&tasks).Error
		if err != nil {
			log.Printf("TaskScheduler:  %v", err)
			time.Sleep(time.Second*10)
			continue
		}

		// Подготавливаем отправку
		for i := range tasks {
			// 1. Ставим статус в работе
			if err := tasks[i].SetStatus(WorkStatusPending); err != nil {
				log.Printf("taskWorker: Ошибка установки статуса Pending задачи [%v]: %v", tasks[i].Id, err)
				continue
			} 
			if err = tasks[i].Execute(); err != nil {

				// Если задача провалена
				if err := tasks[i].SetStatus(WorkStatusFailed); err != nil {
					log.Printf("taskWorker: Ошибка установки статуса Failed задачи [%v]: %v", tasks[i].Id, err)
				}
				continue
			}

			// 2. Если задача выполнена
			if err := tasks[i].SetStatus(WorkStatusCompleted); err != nil {
				log.Printf("taskWorker: Ошибка установки статуса Completed задачи [%v]: %v", tasks[i].Id, err)
				
				// Останавливаем будущий обход до разбирательства
				withoutMistakesTaskWorker = false
			}
			continue

		}

		if len(tasks) > 100 {
			time.Sleep(time.Minute * 2)
			continue
		}
		
		time.Sleep(time.Second * 30) // 2-5s
		continue
	}
}
