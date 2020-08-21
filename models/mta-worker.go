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

		// Буфер MTA-сервера должен вместить нашу пачку писем на отправку
		if (MTACapChannel() - MTALenChannel()) < uint(workflowsOneTick) {
			time.Sleep(time.Second * 2) // <<< надо вообще настроить этот параметр
			log.Printf("Не хватает места в mta-буфере, сободных мест: %v\n", MTACapChannel() - MTALenChannel())
			continue
		}
		
		workflows := make([]MTAWorkflow,0)

		// Собираем по {workflowsOneTick} задач на отправку
		err := db.Model(&MTAWorkflow{}).
			Joins("LEFT JOIN email_queues ON email_queues.id = mta_workflows.owner_id").
			Joins("LEFT JOIN email_notifications ON email_notifications.id = mta_workflows.owner_id").
			Joins("LEFT JOIN email_campaigns ON email_campaigns.id = mta_workflows.owner_id").
			Select("email_queues.status,email_notifications.status, email_campaigns.status, mta_workflows.*").
			Where("mta_workflows.expected_time_start <= ? AND (email_queues.status = ? OR email_notifications.status = ? OR email_campaigns.status = ?)",
				time.Now().UTC(), WorkStatusActive,WorkStatusActive,WorkStatusActive).
			Limit(workflowsOneTick).Find(&workflows).Error
		if err != nil {
			// log.Printf("MTAWorkflow:  %v", err)
			time.Sleep(time.Second*10)
			continue
		}

		// Готовим пакеты на отправку в одном потоке. 
		for i := range workflows {
			if err = workflows[i].Execute(); err != nil {
				// невозможно почему-то отправить письмо
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

		// Чуть отдыхаем перед новым проходом
		time.Sleep(time.Millisecond * 1000)
		continue
	}
}
