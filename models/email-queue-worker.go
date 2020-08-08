package models

import (
	"fmt"
	"log"
	"time"
)

// обходит email_queue_workflows каждые N секунд
// Считывает по 100 актуальных задач
// Получает данные для отправки писем
// Добавляет письма в MTA-server

var maxAttempts = uint(3)

func init() {
	go emailQueueWorker()
}

func emailQueueWorker() {
	fmt.Println("Запускаем emailQueueWorker ")
	for {
		if db == nil {
			time.Sleep(time.Millisecond*2000)
			continue
		}
		
		workflows := make([]EmailQueueWorkflow,0)

		// Получаем задачи у которых серия запущена
		err := db.Model(&EmailQueueWorkflow{}).
			Joins("LEFT JOIN email_queues ON email_queues.id = email_queue_workflows.email_queue_id").
			Select("email_queues.enabled, email_queue_workflows.*").
			Where("email_queues.enabled = 'true' AND email_queue_workflows.expected_time_start <= ?", time.Now().UTC()).Limit(100).Find(&workflows).Error
		if err != nil {
			log.Printf("emailQueueWorker:  %v", err)
			time.Sleep(time.Second*10)
			continue
		}

		// Подготавливаем отправку
		for i := range workflows {
			if err = workflows[i].Execute(); err != nil {
				// fmt.Println("Попыток: ", workflows[i].NumberOfAttempts)
				// Если слишком много попыток
				if workflows[i].NumberOfAttempts + 1 > maxAttempts {
					_ = workflows[i].delete()
				}
			}
		}

		// "Pause" by worflows volume
		if len(workflows) > 1000 {
			time.Sleep(time.Second*120)
			continue
		}
		if len(workflows) > 100 {
			time.Sleep(time.Second*10)
			continue
		}
		// else
		time.Sleep(time.Second*5) // 2-5s
		continue
	}
}
