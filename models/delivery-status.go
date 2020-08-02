package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"time"
)

type DeliveryStatus struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// new, canceled, ...
	Code	string 	`json:"code" gorm:"type:varchar(32);unique;not null;"`

	// new, agreement, equipment, delivery, completed, canceled
	Group	string 	`json:"group" gorm:"type:varchar(32);"` // <<< является так же ключом для понимания общего состояния процесса (completed / canceled)
	GroupName	string 	`json:"group" gorm:"type:varchar(128);"`

	// 'Новый заказ', 'Передан в комплектацию', 'Отменен'
	Name	string `json:"name" gorm:"type:varchar(128);"`

	Description string 	`json:"description" gorm:"type:varchar(255);"` // Описание назначения канала

	CreatedAt 		time.Time `json:"createdAt"`
	UpdatedAt 		time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (deliveryStatus DeliveryStatus) GetId() uint { return deliveryStatus.Id }
func (deliveryStatus *DeliveryStatus) setId(id uint) { deliveryStatus.Id = id }
func (deliveryStatus *DeliveryStatus) setPublicId(id uint) { }
func (deliveryStatus DeliveryStatus) GetAccountId() uint { return deliveryStatus.AccountId }
func (deliveryStatus *DeliveryStatus) setAccountId(id uint) { deliveryStatus.AccountId = id }
func (deliveryStatus DeliveryStatus) SystemEntity() bool { return deliveryStatus.AccountId == 1 }

// ############# Entity interface #############

func (DeliveryStatus) PgSqlCreate() {
	db.AutoMigrate(&DeliveryStatus{})
	db.Model(&DeliveryStatus{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")


	mainAccount, err := GetMainAccount()
	if err != nil {
		log.Println("Не удалось найти главный аккаунт для DeliveryStatus")
	}

	db.Delete(&DeliveryStatus{})
	deliveryStatuses := []DeliveryStatus{
		// new, agreement, delivery, completed, canceled
		{Name: "Новая доставка", 			Code: "new", 				Group:"new", 			GroupName:"Необработанный заказ",	Description: "Необработанный заказ, первоначальный статус заказа на доставку."},

		{Name: "Доставка подтверждена", 	Code: "agreement_completed",Group: "agreement", 	GroupName:"Согласование", Description: "Доставка согласована."},
		{Name: "Предложена замена", 		Code: "agreement_change", 	Group: "agreement", 	GroupName:"Согласование", Description: "Доставка перенесена."},

		{Name: "В процессе доставки", 	Code: "delivery", 	Group: "delivery", 	GroupName:"Доставка", Description: "Заказ в процессе доставки"},

		{Name: "Заказ передан клиенту", 	Code: "completed", Group: "completed", GroupName:"Выполнение", Description: "Доставка выполнена (завершена)"},

		{Name: "Недозвон", 			Code: "canceled_call", 			Group: "canceled", GroupName:"Отмена", Description: "Заказ отменен из-за того что клиент не отвечает."},
		{Name: "Не устроило качество товара", 	Code: "canceled_quality", 	Group: "canceled", GroupName:"Отмена", Description: "Заказ отменен"},
		{Name: "Отменен", 			Code: "canceled_any", 			Group: "canceled", GroupName:"Отмена", Description: "Заказ отменен"},
	}
	for _,v := range deliveryStatuses {
		_, err = mainAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func (deliveryStatus *DeliveryStatus) BeforeCreate(scope *gorm.Scope) error {
	deliveryStatus.Id = 0
	return nil
}


// ######### CRUD Functions ############
func (deliveryStatus DeliveryStatus) create() (Entity, error)  {

	_deliveryStatus := deliveryStatus

	if err := db.Create(&_deliveryStatus).Error; err != nil {
		return nil, err
	}

	var newItem Entity = &_deliveryStatus

	return newItem, nil
}

func (DeliveryStatus) get(id uint) (Entity, error) {

	var deliveryStatus DeliveryStatus

	err := db.First(&deliveryStatus, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryStatus, nil
}
func (deliveryStatus *DeliveryStatus) load() error {

	err := db.First(deliveryStatus, deliveryStatus.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (deliveryStatus *DeliveryStatus) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (DeliveryStatus) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return DeliveryStatus{}.getPaginationList(accountId, 0, 100, sortBy, "")
}
func (DeliveryStatus) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	deliveryStatuses := make([]DeliveryStatus,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := db.Model(&DeliveryStatus{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Find(&deliveryStatuses, "name ILIKE ? OR description ILIKE ?", search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryStatus{}).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&DeliveryStatus{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Find(&deliveryStatuses).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryStatus{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryStatuses))
	for i,_ := range deliveryStatuses {
		entities[i] = &deliveryStatuses[i]
	}

	return entities, total, nil
}

func (deliveryStatus *DeliveryStatus) update(input map[string]interface{}) error {

	// work!!!
	if err := db.Set("gorm:association_autoupdate", false).Model(deliveryStatus).Omit("id", "account_id","created_at").
		Updates(input).Error; err != nil {
		return err
	}

	err := db.First(deliveryStatus, deliveryStatus.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (deliveryStatus *DeliveryStatus) delete () error {
	return db.Model(DeliveryStatus{}).Where("id = ?", deliveryStatus.Id).Delete(deliveryStatus).Error
}
// ######### END CRUD Functions ############
