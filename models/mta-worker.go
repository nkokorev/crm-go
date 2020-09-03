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
const 	workflowsOneTick = 20
var 	_beforeCompletedCheck = uint(0) // при достижении 0 вызывается проверка кампаний, для завершения статуса

func mtaWorker() {

	// Разминка перед стартом
	time.Sleep(time.Second*3)
	fmt.Printf("Start mtaWorker: %v\n", time.Now().UTC())

	var slInNextTick = time.Millisecond * 500


	for {
		if db == nil {
			time.Sleep(time.Second*2)
			continue
		}

		// Буфер MTA-сервера должен вместить нашу пачку писем на отправку
		if (MTACapChannel() - MTALenChannel()) < uint(workflowsOneTick) {
			time.Sleep(time.Second * 1) // <<< надо вообще настроить этот параметр
			// log.Printf("mtaWorker: нехватка mta-буфера, сободных мест: %v\n", MTACapChannel() - MTALenChannel())
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
			log.Printf("Error: db.Model(&MTAWorkflow{}): %v\n",err)
			time.Sleep(time.Second*10)
			continue
		}

		// корректируем скорость выборки задач
		if len(workflows) > 10 {
			slInNextTick = time.Millisecond * 200
		} else {
			slInNextTick = time.Millisecond * 1000
		}

		// Готовим пакеты на отправку в одном потоке.
		for i := range workflows {
			if err = workflows[i].Execute(); err != nil {

				// 1. Удаляем задачу т.к. невозможно почему-то отправить письмо
				if err = workflows[i].delete(); err != nil {
					log.Printf("Неудачное удаление workflows[%v]: %v", i, err)
				}

			} else {

				// 1. удаляем задачу по отправке т.к. считаем, что она выполнена.
				if workflows[i].OwnerType != EmailSenderQueue {

					if err = workflows[i].delete(); err != nil {
						log.Printf("Неудачное удаление workflows[%v]: %v", i, err)
					}

				}

				// 2. Добавляем в пул задач на проверку статуса "Выполнено" (возможны другие статусы Failed из-за проблем с отправкой)
				// todo: create system Task if not exist for check remains workflows
				if workflows[i].OwnerType != EmailSenderCampaign {
					// go MTAWorkflow{}.checkActiveEmailCampaign()
				}

			}
		}

		// Проверяем не пора ли завершить какую кампанию
		MTAWorkflow{}.checkActiveEmailCampaign()

		time.Sleep(slInNextTick)
		continue
	}
}

// Устанавливает WorkStatusCompleted завершенным кампаниям
func (MTAWorkflow) checkActiveEmailCampaign() {

	// Check counter before... <>
	if _beforeCompletedCheck > 0 {
		_beforeCompletedCheck--
		return
	}
	
	_beforeCompletedCheck = 10

	campaigns := make([]EmailCampaign,0)

	err := db.Model(&EmailCampaign{}).
		Joins("LEFT JOIN mta_workflows ON mta_workflows.owner_id = email_campaigns.id AND mta_workflows.owner_type = 'email_campaigns'").
		// Select("COUNT(*) AS recipients FROM mta_workflows WHERE owner_type = 'email_campaigns' ").
		Select("mta_workflows.id, email_campaigns.*").
		Where("email_campaigns.status = ? AND mta_workflows.id IS NULL", WorkStatusActive).
		Find(&campaigns).Error
	if err != nil {
		log.Printf("error: %v\n", err.Error())
		return
	}
	
	// log.Printf("Campaigns: %v\n", len(campaigns))

	// Переводим статус найденных кампаний в 'Completed'
	if len (campaigns) > 0 {
		for i := range campaigns {
			_ = campaigns[i].SetCompletedStatus()
		}
	}
}
