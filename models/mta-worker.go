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

func init() {
	go mtaWorker()
}

// Объем выборки задач за один цикл mtaWorker
const workflowsOneTick = 50

func mtaWorker() {

	// Разминка перед стартом
	time.Sleep(time.Second*3)

	fmt.Printf("Start mtaWorker: %v\n", time.Now().UTC())

	for {
		if db == nil {
			time.Sleep(time.Second*2)
			continue
		}

		// Проверяем, что свободная часть буфера по отправке готова вместить наш объем писем
		if (MTACapChannel() - MTALenChannel()) < uint(workflowsOneTick) {
			time.Sleep(time.Second * 5) // <<< надо вообще настроить этот параметр
			log.Printf("Не хватает места в mta-буфере, сободных мест: %v\n", MTACapChannel() - MTALenChannel())
			continue
		}
		
		workflows := make([]MTAWorkflow,0)

		// Собираем по 50 задач на отправку
		err := db.Model(&MTAWorkflow{}).
			Joins("LEFT JOIN email_queues ON email_queues.id = mta_workflows.owner_id").
			Joins("LEFT JOIN email_notifications ON email_notifications.id = mta_workflows.owner_id").
			Joins("LEFT JOIN email_campaigns ON email_campaigns.id = mta_workflows.owner_id").
			Select("email_queues.enabled,email_notifications.enabled, email_campaigns.enabled, mta_workflows.*").
			Where("mta_workflows.expected_time_start <= ? AND (email_queues.enabled = 'true' OR email_notifications.enabled = 'true' OR email_campaigns.enabled = 'true')", time.Now().UTC()).
			Limit(workflowsOneTick).Find(&workflows).Error
		if err != nil {
			// log.Printf("MTAWorkflow:  %v", err)
			time.Sleep(time.Second*10)
			continue
		}

		// Готовим пакеты на отправку в одном потоке. 
		for i := range workflows {
			if err = workflows[i].Execute(); err != nil {
				// delete задача
				if err = workflows[i].delete(); err != nil {
					log.Printf("Неудачное удаление workflows[%v]: %v", i, err)
				}
			} else {
				// удаляем задачу в этом же цикле, если это не серия писем.
				if workflows[i].OwnerType != EmailSenderQueue {
					if err = workflows[i].delete(); err != nil {
						log.Printf("Неудачное удаление workflows[%v]: %v", i, err)
					}
				}
			}
		}

		// Чуть отдыхаем
		time.Sleep(time.Millisecond * 2000)
		continue
	}
}
