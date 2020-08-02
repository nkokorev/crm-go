package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type DeliveryPickup struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"-" gorm:"index;not null"` // аккаунт-владелец ключа
	WebSiteId	uint 	`json:"webSiteId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится

	Code 		string	`json:"code" gorm:"type:varchar(16);default:'pickup';"` // Для идентификации во фронтенде
	Type 		string	`json:"type" gorm:"type:varchar(32);default:'delivery_pickups';"` // Для идентификации

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

	// загружаемый интерфейс
	PaymentMethods		[]PaymentMethod `json:"paymentMethods" gorm:"-"`

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
func (deliveryPickup *DeliveryPickup) setPublicId(id uint) { }
func (deliveryPickup DeliveryPickup) GetAccountId() uint { return deliveryPickup.AccountId }
func (deliveryPickup DeliveryPickup) GetWebSiteId() uint { return deliveryPickup.WebSiteId }
func (deliveryPickup *DeliveryPickup) setAccountId(id uint) { deliveryPickup.AccountId = id }
func (deliveryPickup *DeliveryPickup) setWebSiteId(webSiteId uint) { deliveryPickup.WebSiteId = webSiteId }
func (DeliveryPickup) SystemEntity() bool { return false }

func (deliveryPickup DeliveryPickup) GetCode() string {
	return deliveryPickup.Code
}
func (deliveryPickup DeliveryPickup) GetType() string {
	return deliveryPickup.Type
}
func (deliveryPickup DeliveryPickup) GetPaymentSubject() PaymentSubject {
	return deliveryPickup.PaymentSubject
}
// ############# Entity interface #############


// ###### GORM Functional #######
func (deliveryPickup *DeliveryPickup) BeforeCreate(scope *gorm.Scope) error {
	deliveryPickup.Id = 0
	return nil
}
func (deliveryPickup *DeliveryPickup) AfterFind() (err error) {

	// Get ALL Payment Methods
	methods, err := GetPaymentMethodsByDelivery(deliveryPickup)
	if err != nil { return err }
	deliveryPickup.PaymentMethods = methods

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

	err := db.Preload("PaymentSubject").Preload("VatCode").First(&deliveryPickup, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryPickup, nil
}
func (deliveryPickup *DeliveryPickup) load() error {

	err := db.Preload("PaymentSubject").Preload("VatCode").First(deliveryPickup, deliveryPickup.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (deliveryPickup *DeliveryPickup) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (DeliveryPickup) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return DeliveryPickup{}.getPaginationList(accountId, 0, 100, sortBy, "")
}

func (DeliveryPickup) getListByShop(accountId, websiteId uint) ([]DeliveryPickup, error) {

	deliveryPickups := make([]DeliveryPickup,0)

	err := DeliveryPickup{}.GetPreloadDb(false,false).
		Limit(100).Where( "account_id = ? AND web_site_id = ?", accountId, websiteId).
		Find(&deliveryPickups).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}

	return deliveryPickups, nil
}

func (DeliveryPickup) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	deliveryPickups := make([]DeliveryPickup,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := DeliveryPickup{}.GetPreloadDb(false,false).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

		err := DeliveryPickup{}.GetPreloadDb(false,false).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
		Preload("PaymentSubject").Preload("VatCode").
		Omit("id", "account_id").Updates(input).Error
}
func (deliveryPickup *DeliveryPickup) delete () error {
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

func (deliveryPickup DeliveryPickup) AppendPaymentMethods(paymentMethods []PaymentMethod) error  {
/*	if err := db.Model(&deliveryPickup).Association("PaymentOptions").Append(paymentMethods).Error; err != nil {
		return err
	}*/

	return nil
}
func (deliveryPickup DeliveryPickup) RemovePaymentMethods(paymentMethods []PaymentMethod) error  {
	/*if err := db.Model(&deliveryPickup).Association("PaymentOptions").Delete(paymentMethods).Error; err != nil {
		return err
	}*/

	return nil
}
func (deliveryPickup DeliveryPickup) ExistPaymentMethod(paymentMethod PaymentMethod) bool  {
	return true
	// return db.Model(&deliveryPickup).Where("payment_options.id = ?", paymentOptions.Id).Association("PaymentOptions").Find(&PaymentOption{}).Count() > 0
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

func (deliveryPickup DeliveryPickup) GetPreloadDb(autoUpdate bool, getModel bool) *gorm.DB {
	_db := db

	if autoUpdate { _db.Set("gorm:association_autoupdate", false) }
	if getModel { _db.Model(&deliveryPickup) }

	return _db.Preload("PaymentSubject").Preload("VatCode")
}