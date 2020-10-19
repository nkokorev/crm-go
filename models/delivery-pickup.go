package models

import (
	"errors"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)
type DeliveryPickup struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint	`json:"-" gorm:"index;not null"` // аккаунт-владелец ключа
	WebSiteId	uint 	`json:"web_site_id" gorm:"type:int;index;"` // магазин, к которому относится

	Code 		string	`json:"code" gorm:"type:varchar(16);default:'pickup';"` // Для идентификации во фронтенде
	Type 		string	`json:"type" gorm:"type:varchar(32);default:'delivery_pickups';"` // Для идентификации

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // активен ли способ доставки
	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Самовывоз со основного склада"
	Price 		float64 `json:"price" gorm:"type:numeric;default:0"` // стоимость доставки при самовывозе (может быть и не 0, если через сервис)

	MaxWeight 	float64 `json:"max_weight" gorm:"type:numeric;default:200"` // максимальная масса в кг

	AddressRequired		bool	`json:"address_required" gorm:"type:bool;default:false"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postal_code_required" gorm:"type:bool;default:false"` // Требуется ли индекс в адресе доставки

	// Признак предмета расчета
	PaymentSubjectId	*uint	`json:"payment_subject_id" gorm:"type:int;default:1;"`//
	PaymentSubject 		PaymentSubject `json:"payment_subject"`

	VatCodeId	uint	`json:"vat_code_id" gorm:"type:int;not null;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vat_code"`

	// загружаемый интерфейс
	PaymentMethods		[]PaymentMethod `json:"payment_methods" gorm:"-"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (DeliveryPickup) PgSqlCreate() {
	db.Migrator().CreateTable(&DeliveryPickup{})
	
	// db.Model(&DeliveryPickup{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&DeliveryPickup{}).AddForeignKey("payment_subject_id", "payment_subjects(id)", "RESTRICT", "CASCADE")
	// db.Model(&DeliveryPickup{}).AddForeignKey("vat_code_id", "vat_codes(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE delivery_pickups " +
		"ADD CONSTRAINT delivery_pickups_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT delivery_pickups_payment_subject_id_fkey FOREIGN KEY (payment_subject_id) REFERENCES payment_subjects(id) ON DELETE RESTRICT ON UPDATE CASCADE," +
		"ADD CONSTRAINT delivery_pickups_vat_code_id_fkey FOREIGN KEY (vat_code_id) REFERENCES vat_codes(id) ON DELETE RESTRICT ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
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
func (deliveryPickup *DeliveryPickup) BeforeCreate(tx *gorm.DB) error {
	deliveryPickup.Id = 0
	return nil
}
func (deliveryPickup *DeliveryPickup) AfterFind(tx *gorm.DB) (err error) {

	// Get ALL Payment Methods
	methods, err := GetPaymentMethodsByDelivery(deliveryPickup)
	if err != nil { return err }
	deliveryPickup.PaymentMethods = methods

	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############
func (deliveryPickup DeliveryPickup) create() (Entity, error)  {
	_item := deliveryPickup
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (DeliveryPickup) get(id uint, preloads []string) (Entity, error) {

	var item DeliveryPickup

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (deliveryPickup *DeliveryPickup) load(preloads []string) error {
	if deliveryPickup.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить DeliveryPickup - не указан  Id"}
	}

	err := deliveryPickup.GetPreloadDb(false, false, preloads).First(deliveryPickup, deliveryPickup.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (deliveryPickup *DeliveryPickup) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (DeliveryPickup) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return DeliveryPickup{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (DeliveryPickup) getListByShop(accountId, websiteId uint) ([]DeliveryPickup, error) {

	deliveryPickups := make([]DeliveryPickup,0)

	err := (&DeliveryPickup{}).GetPreloadDb(false,false, []string{"PaymentMethods"}).
		Limit(100).Where( "account_id = ? AND web_site_id = ?", accountId, websiteId).
		Find(&deliveryPickups).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}

	return deliveryPickups, nil
}
func (DeliveryPickup) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	deliveryPickups := make([]DeliveryPickup,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&DeliveryPickup{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryPickups, "name ILIKE ? OR code ILIKE ? OR price ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&DeliveryPickup{}).GetPreloadDb(false,false, nil).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR price ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&DeliveryPickup{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryPickups).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&DeliveryPickup{}).GetPreloadDb(false,false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryPickups))
	for i := range deliveryPickups {
		entities[i] = &deliveryPickups[i]
	}

	return entities, total, nil
}
func (deliveryPickup *DeliveryPickup) update(input map[string]interface{}, preloads []string) error {
	delete(input,"payment_subject")
	delete(input,"vat_code")
	delete(input,"payment_methods")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"web_site_id","payment_subject_id","vat_code_id"}); err != nil {
		return err
	}

	if err := deliveryPickup.GetPreloadDb(false, false, nil).Where("id = ?", deliveryPickup.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := deliveryPickup.GetPreloadDb(false,false, preloads).First(deliveryPickup, deliveryPickup.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (deliveryPickup *DeliveryPickup) delete () error {
	return deliveryPickup.GetPreloadDb(true,false,nil).Where("id = ?", deliveryPickup.Id).Delete(deliveryPickup).Error
}

// ########## End of CRUD Entity interface ###########

func (deliveryPickup DeliveryPickup) GetName () string {
	return deliveryPickup.Name
}
func (deliveryPickup DeliveryPickup) GetVatCode () (*VatCode, error) {
	if err := deliveryPickup.load([]string{"VatCode"}); err != nil {
		return nil, err
	}
	return &deliveryPickup.VatCode, nil
}
func (deliveryPickup DeliveryPickup) CalculateDelivery(deliveryData DeliveryData, weight float64) (float64, error) {
	return deliveryPickup.Price, nil
	// deliveryData.TotalCost = 0
}
func (deliveryPickup DeliveryPickup) checkMaxWeight(weight float64) error {
	// проверяем максимальную массу:
	if weight > deliveryPickup.MaxWeight {
		return utils.Error{Message: fmt.Sprintf("Превышен максимальный вес посылки в %vкг.", deliveryPickup.MaxWeight)}
	}

	return nil
}

func (deliveryPickup DeliveryPickup) AppendPaymentMethods(paymentMethods []PaymentMethod) error  {
/*	if err := db.Model(&deliveryPickup).Association("PaymentOptions").Append(paymentMethods).Error; err != nil {
		return err
	}*/

	return nil
}
func (deliveryPickup DeliveryPickup) RemovePaymentMethods(paymentMethods []PaymentMethod) error  {
	/*if err := db.Model(&deliveryPickup).Association("PaymentOptions").Delete(&paymentMethods).Error; err != nil {
		return err
	}*/

	return nil
}
func (deliveryPickup DeliveryPickup) ExistPaymentMethod(paymentMethod PaymentMethod) bool  {
	return true
	// return db.Model(&deliveryPickup).Where("payment_options.id = ?", paymentOptions.Id).Association("PaymentOptions").Find(&PaymentOption{}).Count() > 0
}

func (deliveryPickup DeliveryPickup) CreateDeliveryOrder(deliveryData DeliveryData, amount PaymentAmount, order Order) (Entity, error)  {
	status, err := DeliveryStatus{}.GetStatusNew()
	if err != nil { return nil, err}
	customerId := uint(0)
	if order.CustomerId != nil {
		customerId = *order.CustomerId
	}
	deliveryOrder := DeliveryOrder{
		AccountId: deliveryPickup.AccountId,
		OrderId:   &order.Id,
		CustomerId: customerId,
		WebSiteId: order.WebSiteId,
		Code:  deliveryPickup.Code,
		MethodId: deliveryPickup.Id,
		Address: deliveryData.Address,
		PostalCode: deliveryData.PostalCode,
		Amount: amount,
		StatusId: status.Id,
	}

	return deliveryOrder.create()

}

func (deliveryPickup *DeliveryPickup) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(deliveryPickup)
	} else {
		_db = _db.Model(&DeliveryPickup{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Amount","PaymentSubject"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}