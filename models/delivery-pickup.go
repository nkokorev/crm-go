package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type DeliveryPickup struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"-" gorm:"index;not null"` // аккаунт-владелец ключа
	WebSiteId	uint 	`json:"webSiteId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится
	Code 		string	`json:"code" gorm:"type:varchar(16);default:'pickup';"` // Для идентификации во фронтенде

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // активен ли способ доставки
	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Самовывоз со основного склада"
	Price 		float64 `json:"price" gorm:"type:numeric;default:0"` // стоимость доставки при самовывозе (может быть и не 0, если через сервис)

	AddressRequired	bool	`json:"addressRequired" gorm:"type:bool;default:false"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postalCodeRequired" gorm:"type:bool;default:false"` // Требуется ли индекс в адресе доставки

	// Признак предмета расчета
	PaymentSubjectId	uint	`json:"paymentSubjectId" gorm:"type:int;not null;"`//
	PaymentSubject 		PaymentSubject `json:"paymentSubject"`

	VatCodeId	uint	`json:"vatCodeId" gorm:"type:int;not null;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vatCode"`

	// Разрешенные методы оплаты для данного типа доставки
	PaymentOptions	[]PaymentOption `json:"paymentOptions" gorm:"many2many:payment_options_delivery_pickups;preload"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (DeliveryPickup) PgSqlCreate() {
	db.CreateTable(&DeliveryPickup{})
	
	db.Model(&DeliveryPickup{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&DeliveryPickup{}).AddForeignKey("payment_subject_id", "payment_subjects(id)", "CASCADE", "CASCADE")
	db.Model(&DeliveryPickup{}).AddForeignKey("vat_code_id", "vat_codes(id)", "CASCADE", "CASCADE")
}

// ############# Entity interface #############
func (deliveryPickup DeliveryPickup) GetId() uint { return deliveryPickup.Id }
func (deliveryPickup *DeliveryPickup) setId(id uint) { deliveryPickup.Id = id }
func (deliveryPickup DeliveryPickup) GetAccountId() uint { return deliveryPickup.AccountId }
func (deliveryPickup *DeliveryPickup) setAccountId(id uint) { deliveryPickup.AccountId = id }
func (deliveryPickup *DeliveryPickup) setShopId(webSiteId uint) { deliveryPickup.WebSiteId = webSiteId }
func (DeliveryPickup) SystemEntity() bool { return false }

func (deliveryPickup DeliveryPickup) GetCode() string {
	return deliveryPickup.Code
}
// ############# Entity interface #############


// ###### GORM Functional #######
func (deliveryPickup *DeliveryPickup) BeforeCreate(scope *gorm.Scope) error {
	deliveryPickup.Id = 0
	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (deliveryPickup DeliveryPickup) create() (Entity, error)  {
	dp := deliveryPickup
	
	if err := db.Create(&dp).Error; err != nil {
		return nil, err
	}
	var entity Entity = &dp

	return entity, nil
}

func (DeliveryPickup) get(id uint) (Entity, error) {

	var deliveryPickup DeliveryPickup

	err := db.Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").First(&deliveryPickup, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryPickup, nil
}

func (deliveryPickup *DeliveryPickup) load() error {

	err := db.Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").First(deliveryPickup, deliveryPickup.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (DeliveryPickup) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return DeliveryPickup{}.getPaginationList(accountId, 0, 100, sortBy, "")
}
func (DeliveryPickup) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	deliveryPickups := make([]DeliveryPickup,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&DeliveryPickup{}).Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryPickups, "name ILIKE ? OR code ILIKE ? OR price ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryPickup{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR price ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&DeliveryPickup{}).Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryPickups).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryPickup{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryPickups))
	for i,_ := range deliveryPickups {
		entities[i] = &deliveryPickups[i]
	}

	return entities, total, nil
}

func (deliveryPickup *DeliveryPickup) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(deliveryPickup).
		Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").
		Omit("id", "account_id").Updates(input).Error
}

func (deliveryPickup DeliveryPickup) delete () error {
	return db.Model(DeliveryPickup{}).Where("id = ?", deliveryPickup.Id).Delete(deliveryPickup).Error
}

// ########## End of CRUD Entity interface ###########

func (deliveryPickup DeliveryPickup) GetName () string {
	return deliveryPickup.Name
}
func (deliveryPickup DeliveryPickup) GetVatCode () VatCode {
	return deliveryPickup.VatCode
}

func (deliveryPickup DeliveryPickup) CalculateDelivery(deliveryData DeliveryData, weight float64) (float64, error) {
	return deliveryPickup.Price, nil
	// deliveryData.TotalCost = 0
}

func (deliveryPickup DeliveryPickup) checkMaxWeight(weight float64) error {
	return nil
}

func (deliveryPickup DeliveryPickup) AppendPaymentOptions(paymentOptions []PaymentOption) error  {
	if err := db.Model(&deliveryPickup).Association("PaymentOptions").Append(paymentOptions).Error; err != nil {
		return err
	}

	return nil
}
func (deliveryPickup DeliveryPickup) RemovePaymentOptions(paymentOptions []PaymentOption) error  {
	if err := db.Model(&deliveryPickup).Association("PaymentOptions").Delete(paymentOptions).Error; err != nil {
		return err
	}

	return nil
}
func (deliveryPickup DeliveryPickup) ExistPaymentOption(paymentOptions PaymentOption) bool  {
	return db.Model(&deliveryPickup).Where("payment_options.id = ?", paymentOptions.Id).Association("PaymentOptions").Find(&PaymentOption{}).Count() > 0
}

func (deliveryPickup DeliveryPickup) CreateDeliveryOrder(deliveryData DeliveryData, amount PaymentAmount, order Order) (Entity, error)  {
	deliveryOrder := DeliveryOrder{
		AccountId: deliveryPickup.AccountId,
		OrderId:   order.Id,
		CustomerId: order.CustomerId,
		WebSiteId: order.WebSiteId,
		Code:  deliveryPickup.Code,
		MethodId: deliveryPickup.Id,
		Address: deliveryData.Address,
		PostalCode: deliveryData.PostalCode,
		Amount: amount,
	}

	return deliveryOrder.create()

}