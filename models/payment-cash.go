package models

import (
	"errors"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"strings"
)

// Условный метод оплаты кешем. Это либо нал, либо перевод на карту.
type PaymentCash struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`

	HashId 		string `json:"hash_id" gorm:"type:varchar(12);uniqueIndex;not null;"` // публичный Id для защиты от спама/парсинга
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	WebSiteId	uint 	`json:"web_site_id" gorm:"type:int;index;"` // магазин, к которому относится

	Code 		string `json:"code" gorm:"type:varchar(32);default:'payment_cash'"`
	Type 		string	`json:"type" gorm:"type:varchar(32);default:'payment_cashes';"` // Для идентификации
	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя интеграции магазина "<name>"
	Label 		string 	`json:"label" gorm:"type:varchar(255);default:'Оплата наличными при получении'"` // 'Оплата при получении'

	// Включен ли данный способ оплаты ??
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"`
	InstantDelivery bool 	`json:"instant_delivery" gorm:"type:bool;default:false"`

	WebSite		WebSite `json:"web_site"`
	// !!! deprecated !!!
	
	// PaymentOption   PaymentOption `gorm:"polymorphic:Owner;"`
}
func (PaymentCash) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&PaymentCash{}); err != nil {log.Fatal(err)}
	// db.Model(&PaymentCash{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&PaymentCash{}).AddForeignKey("web_site_id", "web_sites(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE payment_cashes " +
		"ADD CONSTRAINT payment_cashes_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT payment_cashes_web_site_id_fkey FOREIGN KEY (web_site_id) REFERENCES web_sites(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (paymentCash *PaymentCash) BeforeCreate(tx *gorm.DB) error {
	paymentCash.Id = 0
	paymentCash.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	return nil
}
func (paymentCash *PaymentCash) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(paymentCash)
	} else {
		_db = _db.Model(&PaymentCash{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"WebSite"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ############# Entity interface #############
func (paymentCash PaymentCash) GetId() uint { return paymentCash.Id }
func (paymentCash *PaymentCash) setId(id uint) { paymentCash.Id = id }
func (paymentCash *PaymentCash) setPublicId(id uint) {  }
func (paymentCash PaymentCash) GetAccountId() uint { return paymentCash.AccountId }
func (paymentCash *PaymentCash) setAccountId(id uint) { paymentCash.AccountId = id }
func (PaymentCash) SystemEntity() bool { return false }
// ############# Entity interface #############

// ############# Payment Method interface #############
func (paymentCash PaymentCash) CreatePaymentByOrder(order Order, mode PaymentMode) (*Payment, error) {

	_p := Payment {
		AccountId: paymentCash.AccountId,
		Paid: false,
		Amount: order.Amount,
		IncomeAmount: order.Amount,
		RefundedAmount: PaymentAmount{AccountId: order.AccountId, Value: float64(0), Currency: "RUB"},
		Description:  fmt.Sprintf("Заказ №%v в магазине AiroCliamte", order.Id),  // Видит клиент
		PaymentMethodData: PaymentMethodData{Type: "bank_card"}, // вообще еще вопрос

		// Чтобы понять какой платеж был оплачен!!!
		/*Metadata: postgres.Jsonb{ RawMessage: utils.MapToRawJson(map[string]interface{}{
			"orderId":order.Id,
			"accountId":paymentCash.AccountId,
		})},*/
		Metadata: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
			"orderId":order.Id,
			"accountId":paymentCash.AccountId,
		})),
		SavePaymentMethod: false,
		OwnerId: paymentCash.Id,
		OwnerType: "payment_cashes",
		OrderId: order.Id,
		PaymentMethodId: order.PaymentMethodId,
		PaymentMethodType: order.PaymentMethodType,
	}

	// создаем внутри платеж
	entity, err := _p.create()
	if err != nil {
		return nil, err
	}
	payment := entity.(*Payment)

	return payment, nil
}
func (paymentCash PaymentCash) PrepaymentCheck(payment *Payment, order Order) (*Payment, error){
	 return payment, nil
}
func (paymentCash PaymentCash) GetWebSiteId() uint { return paymentCash.WebSiteId }
func (paymentCash PaymentCash) GetType() string { return "payment_cashes" }
func (paymentCash PaymentCash) GetCode() string { return "payment_cash" }
func (paymentCash PaymentCash) IsInstantDelivery() bool { return paymentCash.InstantDelivery }
// ############# END OF Payment Method interface #############

// ######### CRUD Functions ############
func (paymentCash PaymentCash) create() (Entity, error)  {
	_item := paymentCash
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (PaymentCash) get(id uint, preloads []string) (Entity, error) {
	var item PaymentCash

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (paymentCash *PaymentCash) load(preloads []string) error {
	if paymentCash.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := paymentCash.GetPreloadDb(false, false, preloads).First(paymentCash, paymentCash.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*PaymentCash) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (PaymentCash) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return  PaymentCash{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (PaymentCash) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}, preloads []string) ([]Entity, int64, error) {

	paymentCashs := make([]PaymentCash,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&PaymentCash{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&paymentCashs, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&PaymentCash{}).GetPreloadDb(false,false, nil).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&PaymentCash{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&paymentCashs).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&PaymentCash{}).GetPreloadDb(false,false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(paymentCashs))
	for i := range paymentCashs {
		entities[i] = &paymentCashs[i]
	}

	return entities, total, nil
}
func (paymentCash *PaymentCash) update(input map[string]interface{}, preloads []string) error {

	delete(input,"web_site")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"web_site_id"}); err != nil {
		return err
	}

	if err := paymentCash.GetPreloadDb(false, false, nil).Where("id = ?", paymentCash.Id).Omit("id", "account_id","hash_id").Updates(input).
		Error; err != nil {return err}

	err := paymentCash.GetPreloadDb(false,false, preloads).First(paymentCash, paymentCash.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (paymentCash *PaymentCash) delete () error {
	return paymentCash.GetPreloadDb(true,false, nil).Where("id = ?", paymentCash.Id).Delete(paymentCash).Error
}
// ######### END CRUD Functions ############

// ########## Work function ############


func (PaymentCash) GetListByWebSiteAndDelivery(delivery Delivery) ([]PaymentCash, error) {

	methods := make([]PaymentCash,0)

	err := db.Table("payment_to_delivery").
		Joins("LEFT JOIN payment_cashes ON payment_cashes.id = payment_to_delivery.payment_id AND payment_cashes.type = payment_to_delivery.payment_type").
		Select("payment_to_delivery.account_id, payment_to_delivery.web_site_id, payment_to_delivery.payment_id, payment_to_delivery.payment_type, payment_to_delivery.delivery_id, payment_to_delivery.delivery_type, payment_cashes.*").
		Where("payment_to_delivery.account_id = ? AND payment_to_delivery.web_site_id = ? " +
			"AND payment_to_delivery.delivery_id = ? AND payment_to_delivery.delivery_type = ? " +
			"AND payment_to_delivery.payment_type = ?",
			delivery.GetAccountId(), delivery.GetWebSiteId(), delivery.GetId(), delivery.GetType(), PaymentCash{}.GetType()).Find(&methods).Error

	if err != nil { return nil, err }

	return methods,nil
}