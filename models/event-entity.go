package models

import (
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

// Список событий, на которые можно навесить обработчик EventHandler
type Event struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"account_id" gorm:"type:int;index;not null;"`

	// #### Entity ####

	// Полезная нагрузка события
	data 		map[string]interface{} `json:"data" gorm:"-"`

	// Список пользователей (иногда нужно)
	recipientList 	[]uint `json:"recipient_list" gorm:"-"`
	// target
	target 			interface{} `json:"-" gorm:"-"`
	// mark is aborted
	aborted bool 		`json:"aborted" gorm:"-"`
	/// #### END of Entity ####

	Name		string 	`json:"name" gorm:"type:varchar(128);unique;not null;"`  // 'Пользователь создан'
	Code		string 	`json:"code" gorm:"type:varchar(128);unique;not null;"`  // 'UserCreated'

	// deprecated

	// Доступен ли вызов через API с контекстом. У почти всех системных событий = false
	ExternalCallAvailable 		bool 	`json:"external_call_available" gorm:"type:bool;default:false;"`

	// Получение списка пользователей / данных из API в контексте recipient_list []
	ParsingRecipientList 	bool 	`json:"parsing_recipient_list" gorm:"type:bool;default:false;"`
	ParsingPayload		 	bool 	`json:"parsing_payload" gorm:"type:bool;default:false;"`

	// Данные события: recipient_list[], DataJson{},
	
	Description string 	`json:"description" gorm:"type:varchar(255);"`
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
		{Name: "Пользователь создан", 	Code: "UserCreated", Description: "Создание пользователя в текущем аккаунте. Сам пользователь на момент вызова не имеет доступа к аккаунту (если вообще будет)."},
		{Name: "Пользователь обновлен",Code: "UserUpdated", Description: "Какие-то данные в учетной записи пользователя обновились."},
		{Name: "Пользователь удален", 	Code: "UserDeleted", Description: "Учетная запись пользователя удалена из системы RatusCRM."},

		{Name: "Пользователь добавлен в аккаунт", Code: "UserAppendedToAccount", Description: "Пользователь получил доступ в текущий аккаунт с какой-то конкретно ролью."},
		{Name: "Пользователь удален из аккаунта", Code: "UserRemovedFromAccount", Description: "У пользователя больше нет доступа к вашей системе из-под своей учетной записи."},

		{Name: "Пользователь подписался на рассылки",	Code: "UserSubscribed", Description: "Пользователь подписался на рассылки. Технически, скорее всего, его 'подписали' через API или GUI интерфейсы."},
		{Name: "Пользователь отписался от рассылок", 	Code: "UserUnsubscribed", Description: "Пользователь отписался от всех рассылок, кроме системных уведомлений."},
		{Name: "Пользователь изменил статус подписки", Code: "UserUpdateSubscribeStatus", Description: "У пользователя обновился статус подписки."},

		{Name: "Товар создан", Code: "ProductCreated", Description: "Создан новый товар или услуга."},
		{Name: "Товар обновлен",	Code: "ProductUpdated", Description: "Данные товара или услуга были обновлены. Сюда также входит обновление связанных данных: изображений, описаний, видео."},
		{Name: "Товар удален", Code: "ProductDeleted", Description: "Товар или услуга удалены из системы со всеми связанными данными."},

		{Name: "Карточка товара создана", 	Code: "ProductCardCreated", Description: "Карточка товара создана в системе"},
		{Name: "Карточка товара обновлена",Code: "ProductCardUpdated", Description: "Данные карточки товара успешно обновлены."},
		{Name: "Карточка товара удалена", 	Code: "ProductCardDeleted", Description: "Карточка товара удалена из системы"},

		{Name: "Страница сайта создана", 	Code: "WebPageCreated", Description: "Создан новый раздел, категория или страница на сайте."},
		{Name: "Страница сайта обновлена", Code: "WebPageUpdated", Description: "Данные раздела или категории сайта успешно обновлены."},
		{Name: "Страница сайта удалена", 	Code: "WebPageDeleted", Description: "Раздел сайта или категория удалена из системы"},

		{Name: "Сайт создан", 	Code: "WebSiteCreated", Description: "Создан новый сайт или магазин."},
		{Name: "Сайт обновлен",Code: "WebSiteUpdated", Description: "Персональные данные сайта или магазина были успешно обновлены."},
		{Name: "Сайт удален", 	Code: "WebSiteDeleted", Description: "Сайт или магазин удален из системы."},

		{Name: "Файл создан", 	Code: "StorageCreated", Description: "В системе создан новый файл."},
		{Name: "Файл обновлен",Code: "StorageUpdated", Description: "Какие-то данные файла успешно изменены."},
		{Name: "Файл удален", 	Code: "StorageDeleted", Description: "Файл удален из системы."},

		{Name: "Статья создана", 	Code: "ArticleCreated", Description: "В системе создана новая статья."},
		{Name: "Статья обновлена", Code: "ArticleUpdated", Description: "Какие-то данные статьи были изменены. Учитываются также и смежные данные, вроде изображений и видео."},
		{Name: "Статья удалена", 	Code: "ArticleDeleted", Description: "Статья со смежными данными удалена из системы."},

		{Name: "Заказ создан", 	Code: "OrderCreated", Description: "Создан новый заказ. В контексте глобальный id заказа."},
		{Name: "Заказ обновлен", 	Code: "OrderUpdated", Description: "Какие-то данные заказа были изменены. В контексте глобальный id заказа."},
		{Name: "Заказ удален", 	Code: "OrderDeleted", Description: "Заказ удален из системы. В контексте глобальный id заказа."},
		{Name: "Заказ выполнен", 	Code: "OrderCompleted", Description: "Заказ выполнен успешно. В контексте глобальный id заказа."},
		{Name: "Заказ отменен", 	Code: "OrderCanceled", Description: "Заказ отменен по каким-то причинам. В контексте глобальный id заказа."},

		{Name: "Создано задание на доставку", 	Code: "DeliveryOrderCreated", Description: "В системе зарегистрировано новое задание на доставку. Это может быть и самовывоз и доставка Почтой России."},
		{Name: "Доставка обновлена", 	Code: "DeliveryOrderUpdated", Description: "Какие-то данные по заказу на доставку обновились."},
		{Name: "Доставка согласована", Code: "DeliveryOrderInProcess", Description: "Задание на доставку в процессе доставки."},
		{Name: "Доставка завершена", 	Code: "DeliveryOrderCompleted", Description: "Задание на доставку успешно завершено."},
		{Name: "Доставка отменена",	Code: "DeliveryOrderCanceled", Description: "Задание на доставку отменено по каким-то причинам."},
		{Name: "У доставки обновился статус", 	Code: "DeliveryOrderStatusUpdated", Description: "Задание на доставку обновило свой статус."},
		{Name: "Доставка удалена", Code: "DeliveryOrderDeleted", Description: "Задание на доставку удалено из системы."},

		{Name: "Создан платеж", 	Code: "PaymentCreated", Description: "Создан объект - платеж (payment). В контексте глобальный id доставки."},
		{Name: "Платеж обновлен", 	Code: "PaymentUpdated", Description: "Какие-то данные платежа изменены. В контексте глобальный id заказа."},
		{Name: "Платеж удален", 	Code: "PaymentDeleted", Description: "Объект платеж удален из системы. В контексте глобальный id заказа."},
		{Name: "Платеж оплачен", 	Code: "PaymentCompleted", Description: "Платеж перешел в статус succeeded или помечен как оплаченный. Учитывается любой из видов расчета: нал/безнал. В контексте глобальный id заказа."},
		{Name: "Платеж отменен", 	Code: "PaymentCanceled", Description: "Платеж отменен по каким-то причинам. В контексте глобальный id заказа."},

		{Name: "Производитель создан", 	Code: "ManufacturerCreated", Description: "В системе добавлен производитель."},
		{Name: "Производитель обновлен",Code: "ManufacturerUpdated", Description: "Данные производителя обновлены."},
		{Name: "Производитель удален", 	Code: "ManufacturerDeleted", Description: "Производитель удален из системы."},

		{Name: "Группа тегов товаров создана", 	Code: "ProductTagGroupCreated", Description: "В системе добавлена группа тегов товаров."},
		{Name: "Группа тегов товаров обновлена",Code: "ProductTagGroupUpdated", Description: "Группа тегов товаров обновлена."},
		{Name: "Группа тегов товаров удалена", 	Code: "ProductTagGroupDeleted", Description: "Группа тегов товаров удалена."},

		{Name: "Тег товаров создан", 	Code: "ProductTagCreated", Description: "В системе добавлен тег товаров."},
		{Name: "Тег товаров обновлен",Code: "ProductTagUpdated", Description: "Тег товаров обновлен."},
		{Name: "Тег товаров удален", 	Code: "ProductTagDeleted", Description: "Тег товаров удален."},

		{Name: "Теги товара синхронизированы", 	Code: "ProductSyncProductTags", Description: "Синхронизация тегов товара."},
		{Name: "Категории товара синхронизированы", 	Code: "ProductSyncProductCategories", Description: "Синхронизация категорий товара."},
		
		{Name: "В карточку товара добавлен товар", 	Code: "ProductCardAppendedProduct", Description: "В карточку товара добавлен товар."},
		{Name: "Из карточки товара убран товар", 	Code: "ProductCardRemovedProduct", 	Description: "Из карточки товара убран товар."},
		
		{Name: "В категорию товара добавлен товар", Code: "ProductCategoryAppendedProduct",Description: "В категорию товара добавлен товар."},
		{Name: "Из категории товара убран товар", 	Code: "ProductCategoryRemovedProduct", Description: "Из категории товара убран товар."},

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
	/*var lastIdx sql.NullInt64
	err := db.Model(&Event{}).Where("account_id = ?",  event.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	event.PublicId = 1 + uint(lastIdx.Int64)*/

	return nil
}

// ############# Entity interface #############
func (event Event) GetId() uint                { return event.Id }
func (event *Event) setId(id uint)             { event.Id = id }
func (event *Event) setPublicId(uint) { }
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

	return utils.Error{Message: "Данный объект невозможно загрузить по public ID"}
	/*if event.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Article - не указан  Id"}
	}
	if err := event.GetPreloadDb(false,false, preloads).First(event, "account_id = ? AND public_id = ?", event.AccountId, event.PublicId).Error; err != nil {
		return err
	}*/

	// return nil
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
			Find(&events, "label ILIKE ? OR code ILIKE ? OR description ILIKE ?",search,search,search).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = (&Event{}).GetPreloadDb(false, false, nil).
			Where("account_id IN (?) AND label ILIKE ? OR code ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search,search).
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