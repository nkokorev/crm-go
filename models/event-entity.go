package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Список событий, на которые можно навесить обработчик EventHandler
type Event struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;default:1"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// #### Entity ####
	name string
	// user data.

	// Полезная нагрузка события
	payload map[string]interface{} `json:"payload" gorm:"-"`
	// target
	target interface{} `json:"-" gorm:"-"`
	// mark is aborted
	aborted bool 		`json:"aborted" gorm:"-"`
	/// #### END of Entity ####

	Label		string 	`json:"label" gorm:"type:varchar(255);unique;not null;"`  // 'Пользователь создан'
	Code		string 	`json:"code" gorm:"type:varchar(255);unique;not null;"`  // 'UserCreated'
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:false;"` // Глобальный статус события (вызывать ли его или нет)

	// JsonData, переданное в событие
	// Payload 	map[string]interface{} `json:"payload" gorm:"-"`
	
	// Доступен ли вызов через API с контекстом. У почти всех системных событий = false
	AvailableAPI 	bool 	`json:"available_api" gorm:"type:bool;default:false;"`

	// Получение списка пользователей из API в контексте recipient_list []
	ParsingRecipientList 	bool 	`json:"parsing_recipient_list" gorm:"type:bool;default:false;"`
	ParsingJsonData		 	bool 	`json:"parsing_json_data" gorm:"type:bool;default:false;"`

	// Данные события: recipient_list[], DataJson{},
	
	Description string 	`json:"description" gorm:"type:text;"` // pgsql: text

	CreatedAt 	time.Time `json:"created_at"`
}

func (Event) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&Event{});err != nil {
		log.Fatal(err)
	}
	// db.Model(&Event{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE events ADD CONSTRAINT event_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	mainAccount, err := GetMainAccount()
	if err != nil {
		log.Println("Не удалось найти главный аккаунт для событий")
	}
	events := []Event{
		{Label: "Пользователь создан", 	Code: "UserCreated", Enabled: true, Description: "Создание пользователя в текущем аккаунте. Сам пользователь на момент вызова не имеет доступа к аккаунту (если вообще будет)."},
		{Label: "Пользователь обновлен",Code: "UserUpdated", Enabled: true, Description: "Какие-то данные в учетной записи пользователя обновились."},
		{Label: "Пользователь удален", 	Code: "UserDeleted", Enabled: true, Description: "Учетная запись пользователя удалена из системы RatusCRM."},

		{Label: "Пользователь добавлен в аккаунт", Code: "UserAppendedToAccount", Enabled: true, Description: "Пользователь получил доступ в текущий аккаунт с какой-то конкретно ролью."},
		{Label: "Пользователь удален из аккаунта", Code: "UserRemovedFromAccount", Enabled: true, Description: "У пользователя больше нет доступа к вашей системе из-под своей учетной записи."},

		{Label: "Пользователь подписался на рассылки",	Code: "UserSubscribed", Enabled: true, Description: "Пользователь подписался на рассылки. Технически, скорее всего, его 'подписали' через API или GUI интерфейсы."},
		{Label: "Пользователь отписался от рассылок", 	Code: "UserUnsubscribed", Enabled: true, Description: "Пользователь отписался от всех рассылок, кроме системных уведомлений."},
		{Label: "Пользователь изменил статус подписки", Code: "UserUpdateSubscribeStatus", Enabled: true, Description: "У пользователя обновился статус подписки."},

		{Label: "Товар создан", Code: "ProductCreated", Enabled: true, Description: "Создан новый товар или услуга."},
		{Label: "Товар обновлен",	Code: "ProductUpdated", Enabled: true, Description: "Данные товара или услуга были обновлены. Сюда также входит обновление связанных данных: изображений, описаний, видео."},
		{Label: "Товар удален", Code: "ProductDeleted", Enabled: true, Description: "Товар или услуга удалены из системы со всеми связанными данными."},

		{Label: "Карточка товара создана", 	Code: "ProductCardCreated", Enabled: true, Description: "Карточка товара создана в системе"},
		{Label: "Карточка товара обновлена",Code: "ProductCardUpdated", Enabled: true, Description: "Данные карточки товара успешно обновлены."},
		{Label: "Карточка товара удалена", 	Code: "ProductCardDeleted", Enabled: true, Description: "Карточка товара удалена из системы"},

		{Label: "Страница сайта создана", 	Code: "WebPageCreated", Enabled: true, Description: "Создан новый раздел, категория или страница на сайте."},
		{Label: "Страница сайта обновлена", Code: "WebPageUpdated", Enabled: true, Description: "Данные раздела или категории сайта успешно обновлены."},
		{Label: "Страница сайта удалена", 	Code: "WebPageDeleted", Enabled: true, Description: "Раздел сайта или категория удалена из системы"},

		{Label: "Сайт создан", 	Code: "WebSiteCreated", Enabled: true, Description: "Создан новый сайт или магазин."},
		{Label: "Сайт обновлен",Code: "WebSiteUpdated", Enabled: true, Description: "Персональные данные сайта или магазина были успешно обновлены."},
		{Label: "Сайт удален", 	Code: "WebSiteDeleted", Enabled: true, Description: "Сайт или магазин удален из системы."},

		{Label: "Файл создан", 	Code: "StorageCreated", Enabled: true, Description: "В системе создан новый файл."},
		{Label: "Файл обновлен",Code: "StorageUpdated", Enabled: true, Description: "Какие-то данные файла успешно изменены."},
		{Label: "Файл удален", 	Code: "StorageDeleted", Enabled: true, Description: "Файл удален из системы."},

		{Label: "Статья создана", 	Code: "ArticleCreated", Enabled: true, Description: "В системе создана новая статья."},
		{Label: "Статья обновлена", Code: "ArticleUpdated", Enabled: true, Description: "Какие-то данные статьи были изменены. Учитываются также и смежные данные, вроде изображений и видео."},
		{Label: "Статья удалена", 	Code: "ArticleDeleted", Enabled: true, Description: "Статья со смежными данными удалена из системы."},

		////////////////// new 29.08.2020
		          
		{Label: "Заказ создан", 	Code: "OrderCreated", Enabled: true, Description: "Создан новый заказ. В контексте глобальный id заказа."},
		{Label: "Заказ обновлен", 	Code: "OrderUpdated", Enabled: true, Description: "Какие-то данные заказа были изменены. В контексте глобальный id заказа."},
		{Label: "Заказ удален", 	Code: "OrderDeleted", Enabled: true, Description: "Заказ удален из системы. В контексте глобальный id заказа."},
		{Label: "Заказ выполнен", 	Code: "OrderCompleted", Enabled: true, Description: "Заказ выполнен успешно. В контексте глобальный id заказа."},
		{Label: "Заказ отменен", 	Code: "OrderCanceled", Enabled: true, Description: "Заказ отменен по каким-то причинам. В контексте глобальный id заказа."},

		{Label: "Создано задание на доставку", 	Code: "DeliveryOrderCreated", Enabled: true, Description: "В системе зарегистрировано новое задание на доставку. Это может быть и самовывоз и доставка Почтой России."},
		{Label: "Доставка обновлена", 	Code: "DeliveryOrderUpdated", Enabled: true, Description: "Какие-то данные по заказу на доставку обновились."},
		{Label: "Доставка согласована", Code: "DeliveryOrderInProcess", Enabled: true, Description: "Задание на доставку в процессе доставки."},
		{Label: "Доставка завершена", 	Code: "DeliveryOrderCompleted", Enabled: true, Description: "Задание на доставку успешно завершено."},
		{Label: "Доставка отменена",	Code: "DeliveryOrderCanceled", Enabled: true, Description: "Задание на доставку отменено по каким-то причинам."},
		{Label: "У доставки обновился статус", 	Code: "DeliveryOrderStatusUpdated", Enabled: true, Description: "Задание на доставку обновило свой статус."},
		{Label: "Доставка удалена", Code: "DeliveryOrderDeleted", Enabled: true, Description: "Задание на доставку удалено из системы."},

		{Label: "Создан платеж", 	Code: "PaymentCreated", Enabled: true, Description: "Создан объект - платеж (payment). В контексте глобальный id доставки."},
		{Label: "Платеж обновлен", 	Code: "PaymentUpdated", Enabled: true, Description: "Какие-то данные платежа изменены. В контексте глобальный id заказа."},
		{Label: "Платеж удален", 	Code: "PaymentDeleted", Enabled: true, Description: "Объект платеж удален из системы. В контексте глобальный id заказа."},
		{Label: "Платеж оплачен", 	Code: "PaymentCompleted", Enabled: true, Description: "Платеж перешел в статус succeeded или помечен как оплаченный. Учитывается любой из видов расчета: нал/безнал. В контексте глобальный id заказа."},
		{Label: "Платеж отменен", 	Code: "PaymentCanceled", Enabled: true, Description: "Платеж отменен по каким-то причинам. В контексте глобальный id заказа."},
	}
	for _,v := range events {
		_, err = mainAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (event *Event) BeforeCreate(tx *gorm.DB) error {
	event.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&Event{}).Where("account_id = ?",  event.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	event.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}

// ############# Entity interface #############
func (event Event) GetId() uint                { return event.Id }
func (event *Event) setId(id uint)             { event.Id = id }
func (event *Event) setPublicId(publicId uint) { event.PublicId = publicId }
func (event Event) GetAccountId() uint         { return event.AccountId }
func (event *Event) setAccountId(id uint)      { event.AccountId = id }
func (event Event) SystemEntity() bool         { return event.AccountId == 1 }

// ############# Entity interface #############

func (event Event) create() (Entity, error)  {

	_item := event
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
func (Event) get(id uint, preloads []string) (Entity, error) {
	var item Event

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (event *Event) load(preloads []string) error {

	if event.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Event - не указан  Id"}
	}

	err := event.GetPreloadDb(false, false, preloads).First(event, event.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (event *Event) loadByPublicId(preloads []string) error {
	if event.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Article - не указан  Id"}
	}
	if err := event.GetPreloadDb(false,false, preloads).First(event, "account_id = ? AND public_id = ?", event.AccountId, event.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (Event) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Event{}.getPaginationList(accountId,0,300, sortBy, "",nil,preload)
}

func (Event) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	events := make([]Event,0)
	var total int64

	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&Event{}).GetPreloadDb(false, false, preloads).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&events, "name ILIKE ? OR description ILIKE ?",search,search).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = (&Event{}).GetPreloadDb(false, false, nil).
			Where("account_id IN (?) AND name ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}


	} else {
		err := (&Event{}).GetPreloadDb(false, false, preloads).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&events).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = (&Event{}).GetPreloadDb(false, false, nil).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(events))
	for i := range events {
		entities[i] = &events[i]
	}

	return entities, total, nil
}

func (event *Event) update(input map[string]interface{}, preloads []string) error {

	// delete(input,"amount")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/

	if err := event.GetPreloadDb(false, false, nil).Where("id = ?", event.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := event.GetPreloadDb(false,false, preloads).First(event, event.Id).Error
	if err != nil {
		return err
	}

	// Перезагружаем все события
	go EventListener{}.ReloadEventHandlers()

	return nil
}

func (event *Event) delete () error {
	if err := event.GetPreloadDb(true,false,nil).Where("id = ?", event.Id).Delete(event).Error;err != nil {return err}
	go EventListener{}.ReloadEventHandlers()

	return nil
}

// ########## Work function ############
func (event *Event) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(event)
	} else {
		_db = _db.Model(&Event{})
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