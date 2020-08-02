package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type DeliveryCourier struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"-" gorm:"index;not null"` // аккаунт-владелец ключа
	WebSiteId	uint 	`json:"webSiteId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится
	Code 		string	`json:"code" gorm:"type:varchar(16);default:'courier';"` // Для идентификации во фронтенде
	Type 		string	`json:"type" gorm:"type:varchar(32);default:'delivery_couriers';"` // Для идентификации

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // активен ли способ доставки
	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Курьерская доставка", "Почта России", "Самовывоз"
	Price 		float64 `json:"price" gorm:"type:numeric;default:0"` // стоимость доставки

	MaxWeight 	float64 `json:"maxWeight" gorm:"type:int;default:40"` // максимальная масса в кг

	AddressRequired	bool	`json:"addressRequired" gorm:"type:bool;default:true"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postalCodeRequired" gorm:"type:bool;default:false"` // Требуется ли индекс в адресе доставки

	// Признак предмета расчета
	PaymentSubjectId	uint	`json:"paymentSubjectId" gorm:"type:int;not null;"`//
	PaymentSubject 		PaymentSubject `json:"paymentSubject"`

	VatCodeId	uint	`json:"vatCodeId" gorm:"type:int;not null;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vatCode"`

	// загружаемый интерфейс
	PaymentMethods		[]PaymentMethod `json:"paymentMethods" gorm:"-"`

	// Список вариантов оплат для указанного магазина. {shopId:}
	// PaymentMethodList 	postgres.Jsonb 	`json:"paymentMethodList" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (DeliveryCourier) PgSqlCreate() {
	db.CreateTable(&DeliveryCourier{})
	db.Model(&DeliveryCourier{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&DeliveryCourier{}).AddForeignKey("payment_subject_id", "payment_subjects(id)", "CASCADE", "CASCADE")
	db.Model(&DeliveryCourier{}).AddForeignKey("vat_code_id", "vat_codes(id)", "CASCADE", "CASCADE")
}

// ############# Entity interface #############
func (deliveryCourier DeliveryCourier) GetId() uint { return deliveryCourier.Id }
func (deliveryCourier *DeliveryCourier) setId(id uint) { deliveryCourier.Id = id }
func (deliveryCourier DeliveryCourier) GetAccountId() uint { return deliveryCourier.AccountId }
func (deliveryCourier DeliveryCourier) GetWebSiteId() uint { return deliveryCourier.WebSiteId }
func (deliveryCourier *DeliveryCourier) setAccountId(id uint) { deliveryCourier.AccountId = id }
func (deliveryCourier *DeliveryCourier) setWebSiteId(webSiteId uint) { deliveryCourier.WebSiteId = webSiteId }
func (DeliveryCourier) SystemEntity() bool { return false }
func (deliveryCourier DeliveryCourier) GetCode() string {
	return deliveryCourier.Code
}
func (deliveryCourier DeliveryCourier) GetType() string {
	return deliveryCourier.Type
}
// ############# Entity interface #############
func (deliveryCourier *DeliveryCourier) BeforeCreate(scope *gorm.Scope) error {
	deliveryCourier.Id = 0
	return nil
}
func (deliveryCourier *DeliveryCourier) AfterFind() (err error) {

	methods, err := GetPaymentMethodsByDelivery(deliveryCourier)
	if err != nil { return err }
	deliveryCourier.PaymentMethods = methods

	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (deliveryCourier DeliveryCourier) create() (Entity, error)  {

	dc := deliveryCourier
	if err := db.Create(&dc).Error; err != nil {
		return nil, err
	}
	var entity Entity = &dc


	return entity, nil
}
func (DeliveryCourier) get(id uint) (Entity, error) {

	var deliveryCourier DeliveryCourier

	err := db.Preload("PaymentSubject").Preload("VatCode").First(&deliveryCourier, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryCourier, nil
}
func (deliveryCourier *DeliveryCourier) load() error {

	err := db.Preload("PaymentSubject").Preload("VatCode").First(deliveryCourier, deliveryCourier.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (DeliveryCourier) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	return DeliveryCourier{}.getPaginationList(accountId, 0, 100, sortBy, "")
}
func (DeliveryCourier) getListByShop(accountId, websiteId uint) ([]DeliveryCourier, error) {

	deliveryCouriers := make([]DeliveryCourier,0)

	err := DeliveryCourier{}.GetPreloadDb(false,false).
		Limit(100).Where( "account_id = ? AND web_site_id = ?", accountId, websiteId).
		Find(&deliveryCouriers).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}

	return deliveryCouriers, nil
}

func (DeliveryCourier) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	deliveryCouriers := make([]DeliveryCourier,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := DeliveryCourier{}.GetPreloadDb(false,false).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryCouriers, "name ILIKE ? OR code ILIKE ? OR price ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryCourier{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR price ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := DeliveryCourier{}.GetPreloadDb(false,false).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryCouriers).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryCourier{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryCouriers))
	for i,_ := range deliveryCouriers {
		entities[i] = &deliveryCouriers[i]
	}

	return entities, total, nil
}
func (deliveryCourier *DeliveryCourier) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(deliveryCourier).
		Preload("PaymentSubject").Preload("VatCode").Omit("id", "account_id").Updates(input).Error
}
func (deliveryCourier *DeliveryCourier) delete () error {
	return db.Model(DeliveryCourier{}).Where("id = ?", deliveryCourier.Id).Delete(deliveryCourier).Error
}

// ########## End of CRUD Entity interface ###########

func (deliveryCourier DeliveryCourier) GetName () string {
	return deliveryCourier.Name
}
func (deliveryCourier DeliveryCourier) GetVatCode () VatCode {
	return deliveryCourier.VatCode
}
func (deliveryCourier DeliveryCourier) GetPaymentSubject() PaymentSubject {	return deliveryCourier.PaymentSubject }

func (deliveryCourier DeliveryCourier) CalculateDelivery(deliveryData DeliveryData, weight float64) (float64, error) {
	return  deliveryCourier.Price, nil
	// deliveryData.TotalCost = deliveryCourier.Price
	// return &deliveryData, nil
}
func (deliveryCourier DeliveryCourier) checkMaxWeight(weight float64) error {
	// проверяем максимальную массу:
	if weight > deliveryCourier.MaxWeight {
		return utils.Error{Message: fmt.Sprintf("Превышен максимальный вес посылки в %vкг.", deliveryCourier.MaxWeight)}
	}

	return nil
}

func (deliveryCourier DeliveryCourier) AppendPaymentMethods(paymentMethods []PaymentMethod) error  {
	return nil
}
func (deliveryCourier DeliveryCourier) RemovePaymentMethods(paymentMethods []PaymentMethod) error  {
	return nil
}
func (deliveryCourier DeliveryCourier) ExistPaymentMethod(paymentMethod PaymentMethod) bool  {
	return true
}

func (deliveryCourier DeliveryCourier) CreateDeliveryOrder(deliveryData DeliveryData, amount PaymentAmount, order Order) (Entity, error)  {
	deliveryOrder := DeliveryOrder{
		AccountId: deliveryCourier.AccountId,
		OrderId:   order.Id,
		CustomerId: order.CustomerId,
		WebSiteId: order.WebSiteId,
		Code:  deliveryCourier.Code,
		MethodId: deliveryCourier.Id,
		Address: deliveryData.Address,
		PostalCode: deliveryData.PostalCode,
		Amount: amount,
	}

	return deliveryOrder.create()

}

func (deliveryCourier DeliveryCourier) GetPreloadDb(autoUpdate bool, getModel bool) *gorm.DB {
	_db := db

	if autoUpdate { _db.Set("gorm:association_autoupdate", false) }
	if getModel { _db.Model(&deliveryCourier) }

	return _db.Preload("PaymentSubject").Preload("VatCode")
}