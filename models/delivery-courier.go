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

type DeliveryCourier struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint	`json:"-" gorm:"index;not null"` // аккаунт-владелец ключа
	WebSiteId	uint 	`json:"web_site_id" gorm:"type:int;index;"` // магазин, к которому относится
	Code 		string	`json:"code" gorm:"type:varchar(16);default:'courier';"` // Для идентификации во фронтенде
	Type 		string	`json:"type" gorm:"type:varchar(32);default:'delivery_couriers';"` // Для идентификации

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // активен ли способ доставки
	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Курьерская доставка", "Почта России", "Самовывоз"
	Price 		float64 `json:"price" gorm:"type:numeric;default:0"` // стоимость доставки

	MaxWeight 	float64 `json:"max_weight" gorm:"type:numeric;default:40"` // максимальная масса в кг

	AddressRequired		bool	`json:"address_required" gorm:"type:bool;default:true"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postal_code_required" gorm:"type:bool;default:false"` // Требуется ли индекс в адресе доставки

	// Признак предмета расчета
	PaymentSubjectId	*uint	`json:"payment_subject_id" gorm:"type:int;not null;"`//
	PaymentSubject 		PaymentSubject `json:"payment_subject"`

	VatCodeId	uint	`json:"vat_code_id" gorm:"type:int;not null;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vat_code"`

	// загружаемый интерфейс
	PaymentMethods		[]PaymentMethod `json:"payment_methods" gorm:"-"`

	// Список вариантов оплат для указанного магазина. {shopId:}
	// PaymentMethodList 	postgres.Jsonb 	`json:"paymentMethodList" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (DeliveryCourier) PgSqlCreate() {
	db.Migrator().CreateTable(&DeliveryCourier{})
	// db.Model(&DeliveryCourier{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&DeliveryCourier{}).AddForeignKey("payment_subject_id", "payment_subjects(id)", "RESTRICT", "CASCADE")
	// db.Model(&DeliveryCourier{}).AddForeignKey("vat_code_id", "vat_codes(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE delivery_couriers " +
		"ADD CONSTRAINT delivery_couriers_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT delivery_couriers_payment_subject_id_fkey FOREIGN KEY (payment_subject_id) REFERENCES payment_subjects(id) ON DELETE RESTRICT ON UPDATE CASCADE," +
		"ADD CONSTRAINT delivery_couriers_vat_code_id_fkey FOREIGN KEY (vat_code_id) REFERENCES vat_codes(id) ON DELETE RESTRICT ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

// ############# Entity interface #############
func (deliveryCourier DeliveryCourier) GetId() uint { return deliveryCourier.Id }
func (deliveryCourier *DeliveryCourier) setId(id uint) { deliveryCourier.Id = id }
func (deliveryCourier *DeliveryCourier) setPublicId(id uint) { }
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
func (deliveryCourier *DeliveryCourier) BeforeCreate(tx *gorm.DB) error {
	deliveryCourier.Id = 0
	return nil
}
func (deliveryCourier *DeliveryCourier) AfterFind(tx *gorm.DB) (err error) {

	methods, err := GetPaymentMethodsByDelivery(deliveryCourier)
	if err != nil { return err }
	deliveryCourier.PaymentMethods = methods

	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (deliveryCourier DeliveryCourier) create() (Entity, error)  {

	_deliveryCourier := deliveryCourier
	if err := db.Create(&_deliveryCourier).Error; err != nil {
		return nil, err
	}

	if err := _deliveryCourier.GetPreloadDb(false,false, nil).First(&_deliveryCourier,_deliveryCourier.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_deliveryCourier

	return entity, nil
}
func (DeliveryCourier) get(id uint, preloads []string) (Entity, error) {

	var deliveryCourier DeliveryCourier

	err := deliveryCourier.GetPreloadDb(false, false, preloads).First(&deliveryCourier, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryCourier, nil
}
func (deliveryCourier *DeliveryCourier) load(preloads []string) error {

	if deliveryCourier.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить DeliveryCourier - не указан  Id"}
	}

	err := deliveryCourier.GetPreloadDb(false, false, preloads).First(deliveryCourier, deliveryCourier.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (deliveryCourier *DeliveryCourier) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (DeliveryCourier) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return DeliveryCourier{}.getPaginationList(accountId, 0, 25, sortBy, "", nil,preload)
}
func (DeliveryCourier) getListByShop(accountId, websiteId uint) ([]DeliveryCourier, error) {

	deliveryCouriers := make([]DeliveryCourier,0)

	err := (&DeliveryCourier{}).GetPreloadDb(false, true, nil).
		Limit(100).Where( "account_id = ? AND web_site_id = ?", accountId, websiteId).
		Find(&deliveryCouriers).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}

	return deliveryCouriers, nil
}

func (DeliveryCourier) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	deliveryCouriers := make([]DeliveryCourier,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&DeliveryCourier{}).GetPreloadDb(false, true, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryCouriers, "name ILIKE ? OR code ILIKE ? OR price ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&DeliveryCourier{}).GetPreloadDb(false, true, nil).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR price ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&DeliveryCourier{}).GetPreloadDb(false, true, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryCouriers).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&DeliveryCourier{}).GetPreloadDb(false, true, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryCouriers))
	for i := range deliveryCouriers {
		entities[i] = &deliveryCouriers[i]
	}

	return entities, total, nil
}
func (deliveryCourier *DeliveryCourier) update(input map[string]interface{}, preloads []string) error {
	delete(input,"payment_subject")
	delete(input,"vat_code")
	delete(input,"payment_methods")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"web_site_id","payment_subject_id","vat_code_id"}); err != nil {
		return err
	}

	if err := deliveryCourier.GetPreloadDb(false, false, nil).Where("id = ?", deliveryCourier.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := deliveryCourier.GetPreloadDb(false,false, preloads).First(deliveryCourier, deliveryCourier.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (deliveryCourier *DeliveryCourier) delete () error {
	return deliveryCourier.GetPreloadDb(true,false,nil).Where("id = ?", deliveryCourier.Id).Delete(deliveryCourier).Error
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
	status, err := DeliveryStatus{}.GetStatusNew()
	if err != nil { return nil, err}

	deliveryOrder := DeliveryOrder{
		AccountId: deliveryCourier.AccountId,
		OrderId:   &order.Id,
		CustomerId: order.CustomerId,
		WebSiteId: order.WebSiteId,
		Code:  deliveryCourier.Code,
		MethodId: deliveryCourier.Id,
		Address: deliveryData.Address,
		PostalCode: deliveryData.PostalCode,
		Amount: amount,
		StatusId: status.Id,
	}

	return deliveryOrder.create()

}

func (deliveryCourier *DeliveryCourier) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(deliveryCourier)
	} else {
		_db = _db.Model(&DeliveryCourier{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"PaymentSubject","VatCode"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}