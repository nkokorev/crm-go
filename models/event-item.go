package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Список событий, на которые можно навесить обработчик EventHandler
type EventItem struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name		string 	`json:"name" gorm:"type:varchar(255);unique;not null;"`  // 'Пользователь создан'
	Code		string 	`json:"code" gorm:"type:varchar(255);unique;not null;"`  // 'UserCreated'
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:false;"` // Глобальный статус события (вызывать ли его или нет)

	Description string 	`json:"description" gorm:"type:text;"` // pgsql: text

	CreatedAt 	time.Time `json:"created_at"`
}

func (EventItem) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&EventItem{});err != nil {
		log.Fatal(err)
	}
	// db.Model(&EventItem{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE event_items ADD CONSTRAINT event_items_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	mainAccount, err := GetMainAccount()
	if err != nil {
		log.Println("Не удалось найти главный аккаунт для событий")
	}
	eventItems := []EventItem{
		{Name: "Пользователь создан", 	Code: "UserCreated", Enabled: true, Description: "Создание пользователя в текущем аккаунте. Сам пользователь на момент вызова не имеет доступа к аккаунту (если вообще будет)."},
		{Name: "Пользователь обновлен", Code: "UserUpdated", Enabled: true, Description: "Какие-то данные в учетной записи пользователя обновились."},
		{Name: "Пользователь удален", 	Code: "UserDeleted", Enabled: true, Description: "Учетная запись пользователя удалена из системы RatusCRM."},

		{Name: "Пользователь добавлен в аккаунт", Code: "UserAppendedToAccount", Enabled: true, Description: "Пользователь получил доступ в текущий аккаунт с какой-то конкретно ролью."},
		{Name: "Пользователь удален из аккаунта", Code: "UserRemovedFromAccount", Enabled: true, Description: "У пользователя больше нет доступа к вашей системе из-под своей учетной записи."},

		{Name: "Пользователь подписался на рассылки", 	Code: "UserSubscribed", Enabled: true, Description: "Пользователь подписался на рассылки. Технически, скорее всего, его 'подписали' через API или GUI интерфейсы."},
		{Name: "Пользователь отписался от рассылок", Code: "UserUnsubscribed", Enabled: true, Description: "Пользователь отписался от всех рассылок, кроме системных уведомлений."},
		{Name: "Пользователь изменил статус подписки", 	Code: "UserUpdateSubscribeStatus", Enabled: true, Description: "У пользователя обновился статус подписки."},

		{Name: "Товар создан", 		Code: "ProductCreated", Enabled: true, Description: "Создан новый товар или услуга."},
		{Name: "Товар обновлен", 	Code: "ProductUpdated", Enabled: true, Description: "Данные товара или услуга были обновлены. Сюда также входит обновление связанных данных: изображений, описаний, видео."},
		{Name: "Товар удален", 		Code: "ProductDeleted", Enabled: true, Description: "Товар или услуга удалены из системы со всеми связанными данными."},

		{Name: "Карточка товара создана", 	Code: "ProductCardCreated", Enabled: true, Description: "Карточка товара создана в системе"},
		{Name: "Карточка товара обновлена", Code: "ProductCardUpdated", Enabled: true, Description: "Данные карточки товара успешно обновлены."},
		{Name: "Карточка товара удалена", 	Code: "ProductCardDeleted", Enabled: true, Description: "Карточка товара удалена из системы"},

		{Name: "Страница сайта создана", 	Code: "WebPageCreated", Enabled: true, Description: "Создан новый раздел, категория или страница на сайте."},
		{Name: "Страница сайта обновлена", 	Code: "WebPageUpdated", Enabled: true, Description: "Данные раздела или категории сайта успешно обновлены."},
		{Name: "Страница сайта удалена", 	Code: "WebPageDeleted", Enabled: true, Description: "Раздел сайта или категория удалена из системы"},

		{Name: "Сайт создан", 	Code: "WebSiteCreated", Enabled: true, Description: "Создан новый сайт или магазин."},
		{Name: "Сайт обновлен", Code: "WebSiteUpdated", Enabled: true, Description: "Персональные данные сайта или магазина были успешно обновлены."},
		{Name: "Сайт удален", 	Code: "WebSiteDeleted", Enabled: true, Description: "Сайт или магазин удален из системы."},

		{Name: "Файл создан", 	Code: "StorageCreated", Enabled: true, Description: "В системе создан новый файл."},
		{Name: "Файл обновлен", Code: "StorageUpdated", Enabled: true, Description: "Какие-то данные файла успешно изменены."},
		{Name: "Файл удален", 	Code: "StorageDeleted", Enabled: true, Description: "Файл удален из системы."},

		{Name: "Статья создана", 	Code: "ArticleCreated", Enabled: true, Description: "В системе создана новая статья."},
		{Name: "Статья обновлена", 	Code: "ArticleUpdated", Enabled: true, Description: "Какие-то данные статьи были изменены. Учитываются также и смежные данные, вроде изображений и видео."},
		{Name: "Статья удалена", 	Code: "ArticleDeleted", Enabled: true, Description: "Статья со смежными данными удалена из системы."},

		////////////////// new 29.08.2020
		          
		{Name: "Заказ создан", 	Code: "OrderCreated", Enabled: true, Description: "Создан новый заказ. В контексте глобальный id заказа."},
		{Name: "Заказ обновлен", 	Code: "OrderUpdated", Enabled: true, Description: "Какие-то данные заказа были изменены. В контексте глобальный id заказа."},
		{Name: "Заказ удален", 	Code: "OrderDeleted", Enabled: true, Description: "Заказ удален из системы. В контексте глобальный id заказа."},
		{Name: "Заказ выполнен", 	Code: "OrderCompleted", Enabled: true, Description: "Заказ выполнен успешно. В контексте глобальный id заказа."},
		{Name: "Заказ отменен", 	Code: "OrderCanceled", Enabled: true, Description: "Заказ отменен по каким-то причинам. В контексте глобальный id заказа."},

		{Name: "Создано задание на доставку", 	Code: "DeliveryOrderCreated", Enabled: true, Description: "В системе зарегистрировано новое задание на доставку. Это может быть и самовывоз и доставка Почтой России."},
		{Name: "Доставка обновлена", Code: "DeliveryOrderUpdated", Enabled: true, Description: "Какие-то данные по заказу на доставку обновились."},
		{Name: "Доставка согласована", 			Code: "DeliveryOrderInProcess", Enabled: true, Description: "Задание на доставку в процессе доставки."},
		{Name: "Доставка завершена", 	Code: "DeliveryOrderCompleted", Enabled: true, Description: "Задание на доставку успешно завершено."},
		{Name: "Доставка отменена", 	Code: "DeliveryOrderCanceled", Enabled: true, Description: "Задание на доставку отменено по каким-то причинам."},
		{Name: "У доставки обновился статус", 	Code: "DeliveryOrderStatusUpdated", Enabled: true, Description: "Задание на доставку обновило свой статус."},
		{Name: "Доставка удалена", 		Code: "DeliveryOrderDeleted", Enabled: true, Description: "Задание на доставку удалено из системы."},

		{Name: "Создан платеж", 	Code: "PaymentCreated", Enabled: true, Description: "Создан объект - платеж (payment). В контексте глобальный id доставки."},
		{Name: "Платеж обновлен", Code: "PaymentUpdated", Enabled: true, Description: "Какие-то данные платежа изменены. В контексте глобальный id заказа."},
		{Name: "Платеж удален", 		Code: "PaymentDeleted", Enabled: true, Description: "Объект платеж удален из системы. В контексте глобальный id заказа."},
		{Name: "Платеж оплачен", 	Code: "PaymentCompleted", Enabled: true, Description: "Платеж перешел в статус succeeded или помечен как оплаченный. Учитывается любой из видов расчета: нал/безнал. В контексте глобальный id заказа."},
		{Name: "Платеж отменен", 	Code: "PaymentCanceled", Enabled: true, Description: "Платеж отменен по каким-то причинам. В контексте глобальный id заказа."},
	}
	for _,v := range eventItems {
		_, err = mainAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (eventItem *EventItem) BeforeCreate(tx *gorm.DB) error {
	eventItem.Id = 0
	return nil
}

// ############# Entity interface #############
func (eventItem EventItem) GetId() uint { return eventItem.Id }
func (eventItem *EventItem) setId(id uint) { eventItem.Id = id }
func (eventItem *EventItem) setPublicId(id uint) {  }
func (eventItem EventItem) GetAccountId() uint { return eventItem.AccountId }
func (eventItem *EventItem) setAccountId(id uint) { eventItem.AccountId = id }
func (eventItem EventItem) SystemEntity() bool { return eventItem.AccountId == 1 }

// ############# Entity interface #############


func (eventItem EventItem) create() (Entity, error)  {

	_item := eventItem
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	// перезагружаем все события
	go EventListener{}.ReloadEventHandlers()

	return entity, nil
}

func (EventItem) get(id uint, preloads []string) (Entity, error) {
	var item EventItem

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (eventItem *EventItem) load(preloads []string) error {

	if eventItem.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EventItem - не указан  Id"}
	}

	err := eventItem.GetPreloadDb(false, false, preloads).First(eventItem, eventItem.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*EventItem) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (EventItem) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return EventItem{}.getPaginationList(accountId,0,300, sortBy, "",nil,preload)
}

func (EventItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	eventItems := make([]EventItem,0)
	var total int64

	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&EventItem{}).GetPreloadDb(false, false, preloads).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&eventItems, "name ILIKE ? OR description ILIKE ?",search,search).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = (&EventItem{}).GetPreloadDb(false, false, nil).
			Where("account_id IN (?) AND name ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}


	} else {
		err := (&EventItem{}).GetPreloadDb(false, false, preloads).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&eventItems).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = (&EventItem{}).GetPreloadDb(false, false, nil).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(eventItems))
	for i := range eventItems {
		entities[i] = &eventItems[i]
	}

	return entities, total, nil
}

func (eventItem *EventItem) update(input map[string]interface{}, preloads []string) error {

	// delete(input,"amount")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/

	if err := eventItem.GetPreloadDb(false, false, nil).Where("id = ?", eventItem.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := eventItem.GetPreloadDb(false,false, preloads).First(eventItem, eventItem.Id).Error
	if err != nil {
		return err
	}

	// Перезагружаем все события
	go EventListener{}.ReloadEventHandlers()

	return nil
}

func (eventItem *EventItem) delete () error {
	if err := eventItem.GetPreloadDb(true,false,nil).Where("id = ?", eventItem.Id).Delete(eventItem).Error;err != nil {return err}
	go EventListener{}.ReloadEventHandlers()

	return nil
}

// ########## Work function ############
func (eventItem *EventItem) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&eventItem)
	} else {
		_db = _db.Model(&EventItem{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}