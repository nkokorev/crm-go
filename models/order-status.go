package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type OrderStatus struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// new, canceled, ...
	Code		string 	`json:"code" gorm:"type:varchar(32);"`

	// new, agreement, equipment, delivery, completed, canceled
	Group		string 	`json:"group" gorm:"type:varchar(32);"`
	GroupName	string 	`json:"group_name" gorm:"type:varchar(128);"`

	// 'Новый заказ', 'Передан в комплектацию', 'Отменен'
	Name		string `json:"name" gorm:"type:varchar(128);"`

	Description string 	`json:"description" gorm:"type:varchar(255);"` // Описание назначения канала

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

// ############# Entity interface #############
func (orderStatus OrderStatus) GetId() uint { return orderStatus.Id }
func (orderStatus *OrderStatus) setId(id uint) { orderStatus.Id = id }
func (orderStatus *OrderStatus) setPublicId(id uint) { orderStatus.Id = id }
func (orderStatus OrderStatus) GetAccountId() uint { return orderStatus.AccountId }
func (orderStatus *OrderStatus) setAccountId(id uint) { orderStatus.AccountId = id }
func (orderStatus OrderStatus) SystemEntity() bool { return orderStatus.AccountId == 1 }

// ############# Entity interface #############

func (OrderStatus) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&OrderStatus{}); err != nil {log.Fatal(err)}
	// db.Model(&OrderStatus{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE order_statuses ADD CONSTRAINT order_statuses_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	mainAccount, err := GetMainAccount()
	if err != nil {
		log.Println("Не удалось найти главный аккаунт для OrderStatus")
	}

	orderStatuses := []OrderStatus{
		// new, agreement, equipment, delivery, completed, canceled
		{Name: "Новый заказ", 				Code: "new", 					Group:"new", 			GroupName:"Необработанные заявки",	Description: "Необработанный заказ, первоначальный статус заказа."},

		{Name: "Заказ подтвержден", 		Code: "agreement_order", 		Group: "agreement", 	GroupName:"Согласование", 	Description: "Заказ подтвержден."},
		{Name: "Предложена замена", 		Code: "agreement_change", 		Group: "agreement", 	GroupName:"Согласование", 	Description: "Предложена замена, в процессе согласования."},
		{Name: "Согласование с клиентом", 	Code: "agreement_approval", 	Group: "agreement", 	GroupName:"Согласование", 	Description: "В процессе согласование с клиентом."},

		{Name: "Комплектуется", 			Code: "equipping", 				Group: "equipment", 	GroupName:"Комплектация", 	Description: "Комплектуется."},
		{Name: "Укомплектован", 			Code: "equipped", 				Group: "equipment", 	GroupName:"Комплектация", 	Description: "Укомплектован"},

		{Name: "Передан в доставку", 		Code: "delivery_sent", 			Group: "delivery", 		GroupName:"Доставка", 		Description: "Заказ передан в доставку"},
		{Name: "Доставляется", 				Code: "delivering", 			Group: "delivery", 		GroupName:"Доставка", 		Description: "Заказ в процессе доставки"},
		{Name: "Доставка перенесена", 		Code: "delivery_rescheduled", 	Group: "delivery", 		GroupName:"Доставка", 		Description: "Доставка перенесена"},

		{Name: "Заказ выполнен", 			Code: "completed", 				Group: "completed", 	GroupName:"Выполнение", 	Description: "Заказ выполнен (завершен)"},

		{Name: "Недозвон", 					Code: "canceled_call", 			Group: "canceled", 		GroupName:"Отмена", 		Description: "Заказ отменен"},
		{Name: "Нет в наличии", 			Code: "canceled_out_of_stock", 	Group: "canceled", 		GroupName:"Отмена", 		Description: "Заказ отменен"},
		{Name: "Не устроила цена", 			Code: "canceled_price", 		Group: "canceled", 		GroupName:"Отмена", 		Description: "Заказ отменен"},
		{Name: "Не устроила доставка", 		Code: "canceled_delivery", 		Group: "canceled", 		GroupName:"Отмена", 		Description: "Заказ отменен"},
		{Name: "Отменен", 					Code: "canceled_any", 			Group: "canceled", 		GroupName:"Отмена", 		Description: "Заказ отменен"},
		{Name: "Спам", 						Code: "spam", 					Group: "canceled", 		GroupName:"Отмена", 		Description: "Спам заказ"},

		{Name: "Резерв предзаказа",			Code: "pre-order-reserve", 		Group: "prepend", 		GroupName:"Формирование", 	Description: "Резерв товара для предзаказа"},
		{Name: "Резерв для отправки",		Code: "reserve-pre-sender", 	Group: "prepend", 		GroupName:"Формирование", 	Description: "Резерв товара для отправки товара или перемещения"},
	}
	for _,v := range orderStatuses {
		_, err = mainAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func (orderStatus *OrderStatus) BeforeCreate(tx *gorm.DB) error {
	orderStatus.Id = 0
	return nil
}
func (orderStatus *OrderStatus) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(orderStatus)
	} else {
		_db = _db.Model(&OrderStatus{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{""})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ######### CRUD Functions ############
func (orderStatus OrderStatus) create() (Entity, error)  {

	_item := orderStatus
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (OrderStatus) get(id uint, preloads []string) (Entity, error) {

	var item OrderStatus

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (orderStatus *OrderStatus) load(preloads []string) error {
	if orderStatus.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := orderStatus.GetPreloadDb(false, false, preloads).First(orderStatus, orderStatus.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*OrderStatus) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (OrderStatus) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return OrderStatus{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (OrderStatus) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	orderStatuses := make([]OrderStatus,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&OrderStatus{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Find(&orderStatuses, "name ILIKE ? OR description ILIKE ?", search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&OrderStatus{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err :=(&OrderStatus{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Find(&orderStatuses).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&OrderStatus{}).GetPreloadDb(false, false, nil).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(orderStatuses))
	for i := range orderStatuses {
		entities[i] = &orderStatuses[i]
	}

	return entities, total, nil
}

func (orderStatus *OrderStatus) update(input map[string]interface{}, preloads []string) error {

	// delete(input,"order")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/

	if err := orderStatus.GetPreloadDb(false, false, nil).Where("id = ?", orderStatus.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := orderStatus.GetPreloadDb(false,false, preloads).First(orderStatus, orderStatus.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (orderStatus *OrderStatus) delete () error {
	return orderStatus.GetPreloadDb(true,false,nil).Where("id = ?", orderStatus.Id).Delete(orderStatus).Error
}
// ######### END CRUD Functions ############

func (OrderStatus) GetCompletedStatus() (OrderStatus, error) {
	var status OrderStatus

	err := db.First(&status, "code = 'completed'").Error
	if err != nil {
		return status, err
	}
	return status, nil
}

func (OrderStatus) GetCanceledAnyStatus() (OrderStatus, error) {
	var status OrderStatus

	err := db.First(&status, "code = 'canceled_any'").Error
	if err != nil {
		return status, err
	}
	return status, nil
}
