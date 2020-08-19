package models

import (
	"log"
	"time"
)

// обходит email_queue_workflows каждые N секунд
// Считывает по 100 актуальных задач
// Получает данные для отправки писем
// Добавляет письма в MTA-server

var maxAttemptsMTAWorker = uint(3)

func init() {
	go mtaWorker()
}

func mtaWorker() {
	
	for {
		if db == nil {
			time.Sleep(time.Millisecond*2000)
			continue
		}
		
		workflows := make([]MTAWorkflow,0)

		// Получаем задачи у которых серия запущена
		err := db.Model(&MTAWorkflow{}).
			Joins("LEFT JOIN email_queues ON email_queues.id = mta_workflows.owner_id").
			Joins("LEFT JOIN email_notifications ON email_notifications.id = mta_workflows.owner_id").
			Select("email_queues.enabled,email_notifications.enabled, mta_workflows.*").
			// Where("email_queues.enabled = 'true' OR  email_notifications.enabled = 'true' AND mta_workflows.expected_time_start <= ?", time.Now().UTC()).Limit(100).Find(&workflows).Error
			// Where("email_queues.enabled = 'true' AND mta_workflows.expected_time_start <= ?", time.Now().UTC()).Limit(100).Find(&workflows).Error
			Where("mta_workflows.expected_time_start <= ? AND (email_queues.enabled = 'true' OR email_notifications.enabled = 'true')", time.Now().UTC()).Limit(100).Find(&workflows).Error
		if err != nil {
			log.Printf("MTAWorkflow:  %v", err)
			time.Sleep(time.Second*10)
			continue
		}


		// Подготавливаем отправку
		for i := range workflows {
			if err = workflows[i].Execute(); err != nil {
				// fmt.Println("Попыток: ", workflows[i].NumberOfAttempts)
				// Если слишком много попыток
				if workflows[i].NumberOfAttempts + 1 > maxAttemptsMTAWorker {
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
