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

var maxAttempts = 3

func init() {
	go emailQueueWorker()
}

func emailQueueWorker() {
	for {
		if db == nil {
			fmt.Println("db is null")
			time.Sleep(time.Millisecond*1000)
			continue
		}
		workflows := make([]EmailQueueWorkflow,0)

		// Получаем задачи у которых серия запущена
		err := db.Model(&EmailQueueWorkflow{}).
			Joins("LEFT JOIN email_queues ON email_queues.id = email_queue_workflows.email_queue_id").
			Select("email_queues.enabled, email_queue_workflows.*").
			Where("email_queues.enabled = 'true' AND email_queue_workflows.expected_time_start <= ?", time.Now()).Limit(100).Find(&workflows).Error
		if err != nil {
			log.Printf("emailQueueWorker:  %v", err)
		}

		fmt.Printf("%v Найдены workflows: %v \n", time.Now(), len(workflows))

		// Подготавливаем отправку
		for _,v := range workflows {
			if err = v.Execute(); err != nil {
				fmt.Println("Err: ", err)
			}
		}

		// "Pause" by worflows volume
		if len(workflows) < 10 {
			time.Sleep(time.Second*10)
			continue
		}

		if len(workflows) < 100 {
			time.Sleep(time.Second*10)
			continue
		}

		if len(workflows) < 1000 {
			time.Sleep(time.Second*120)
			continue
		}

	}
}
