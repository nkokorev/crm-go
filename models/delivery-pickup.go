package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type DeliveryPickup struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint	`json:"-" gorm:"index,not null"` // аккаунт-владелец ключа
	WebSiteID		uint 	`json:"webSiteID" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится
	Code 		string	`json:"code" gorm:"type:varchar(16);default:'pickup';"` // Для идентификации во фронтенде

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // активен ли способ доставки
	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Самовывоз со основного склада"
	Price 		float64 `json:"price" gorm:"type:numeric;default:0"` // стоимость доставки при самовывозе (может быть и не 0, если через сервис)

	AddressRequired	bool	`json:"addressRequired" gorm:"type:bool;default:false"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postalCodeRequired" gorm:"type:bool;default:false"` // Требуется ли индекс в адресе доставки

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (DeliveryPickup) PgSqlCreate() {
	db.CreateTable(&DeliveryPickup{})
	
	db.Model(&DeliveryPickup{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}

// ############# Entity interface #############
func (deliveryPickup DeliveryPickup) GetID() uint { return deliveryPickup.ID }
func (deliveryPickup *DeliveryPickup) setID(id uint) { deliveryPickup.ID = id }
func (deliveryPickup DeliveryPickup) GetAccountID() uint { return deliveryPickup.AccountID }
func (deliveryPickup *DeliveryPickup) setAccountID(id uint) { deliveryPickup.AccountID = id }
func (deliveryPickup *DeliveryPickup) setShopID(webSiteID uint) { deliveryPickup.WebSiteID = webSiteID }
func (DeliveryPickup) SystemEntity() bool { return false }

func (deliveryPickup DeliveryPickup) GetCode() string {
	return deliveryPickup.Code
}
// ############# Entity interface #############


// ###### GORM Functional #######
func (deliveryPickup *DeliveryPickup) BeforeCreate(scope *gorm.Scope) error {
	deliveryPickup.ID = 0
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

	err := db.First(&deliveryPickup, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryPickup, nil
}

func (deliveryPickup *DeliveryPickup) load() error {

	err := db.First(deliveryPickup).Error
	if err != nil {
		return err
	}
	return nil
}

func (DeliveryPickup) getList(accountID uint, sortBy string) ([]Entity, uint, error) {

	deliveryPickups := make([]DeliveryPickup,0)
	var total uint

	err := db.Model(&DeliveryPickup{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountID).
		Find(&deliveryPickups).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&DeliveryPickup{}).Where("account_id = ?", accountID).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryPickups))
	for i,_ := range deliveryPickups {
		entities[i] = &deliveryPickups[i]
	}

	return entities, total, nil
}
func (DeliveryPickup) getPaginationList(accountID uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	deliveryPickups := make([]DeliveryPickup,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&DeliveryPickup{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountID).
			Find(&deliveryPickups, "name ILIKE ? OR code ILIKE ? OR price ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryPickup{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR price ILIKE ?", accountID, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&DeliveryPickup{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountID).
			Find(&deliveryPickups).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryPickup{}).Where("account_id = ?", accountID).Count(&total).Error
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
	return db.Set("gorm:association_autoupdate", false).Model(deliveryPickup).Omit("id", "account_id").Updates(input).Error
}

func (deliveryPickup DeliveryPickup) delete () error {
	return db.Model(DeliveryPickup{}).Where("id = ?", deliveryPickup.ID).Delete(deliveryPickup).Error
}

// ########## End of CRUD Entity interface ###########

func (deliveryPickup DeliveryPickup) GetName () string {
	return "Почта России"
}

func (deliveryPickup DeliveryPickup) CalculateDelivery(deliveryData DeliveryData) (*DeliveryData, error) {
	deliveryData.TotalCost = 0
	return &deliveryData, nil
}

func (deliveryPickup DeliveryPickup) checkMaxWeight(deliveryData DeliveryData) error {
	return nil
}

