package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// Корзины товаров
type CartItem struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа
	OrderId 	uint	`json:"orderId" gorm:"index;not null"` // заказ, к которому относится корзина

	ProductId	uint 	`json:"productId" gorm:"type:int;default:null;"`// Id позиции товара
	Description	string 	`json:"description" gorm:"type:varchar(128);not null;"`
	Quantity	uint	`json:"quantity" gorm:"type:int;not null;"`// число ед. товара

	// Фиксируем стоимость 
	AmountId  	uint			`json:"amountId" gorm:"type:int;not null;"`
	Amount  	PaymentAmount	`json:"amount"`

	// Признак предмета расчета
	PaymentSubjectId	uint	`json:"paymentSubjectId" gorm:"type:int;not null;default:1"`// товар или услуга ? [вид номенклатуры]
	PaymentSubject 		PaymentSubject `json:"paymentSubject"`
	PaymentSubjectYandex	string `json:"payment_subject" gorm:"-"`

	// Признак способа расчета
	PaymentModeId	uint	`json:"paymentModeId" gorm:"type:int;not null;default:1"`//
	PaymentMode 	PaymentMode `json:"paymentMode"`
	PaymentModeYandex 	string `json:"payment_mode" gorm:"-"`

	// Ставка НДС
	VatCode	uint	`json:"vat_code"`

	Product Product `json:"product"`
	Order	Order `json:"-" gorm:"preload:false"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (CartItem) PgSqlCreate() {
	db.CreateTable(&CartItem{})
	db.Model(&CartItem{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&CartItem{}).AddForeignKey("amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
	db.Model(&CartItem{}).AddForeignKey("order_id", "orders(id)", "CASCADE", "CASCADE")
	db.Model(&CartItem{}).AddForeignKey("payment_subject_id", "payment_subjects(id)", "RESTRICT", "CASCADE")
	db.Model(&CartItem{}).AddForeignKey("payment_mode_id", "payment_modes(id)", "RESTRICT", "CASCADE")

}
func (cartItem *CartItem) BeforeCreate(scope *gorm.Scope) error {
	cartItem.Id = 0
	cartItem.Amount.AccountId = cartItem.AccountId
	return nil
}
func (cartItem *CartItem) AfterFind() (err error) {

	cartItem.PaymentSubjectYandex = cartItem.PaymentSubject.Code

	return nil
}

// ############# Entity interface #############
func (cartItem CartItem) GetId() uint { return cartItem.Id }
func (cartItem *CartItem) setId(id uint) { cartItem.Id = id }
func (cartItem *CartItem) setPublicId(publicId uint) { }
func (cartItem CartItem) GetAccountId() uint { return cartItem.AccountId }
func (cartItem *CartItem) setAccountId(id uint) { cartItem.AccountId = id }
func (cartItem CartItem) SystemEntity() bool { return false; }
// ############# End of Entity interface #############


// ######### CRUD Functions ############
func (cartItem CartItem) create() (Entity, error)  {

	// fix id
	cartItem.Amount.AccountId = cartItem.AccountId
	
	_orderChannel := cartItem
	if err := db.Create(&_orderChannel).Error; err != nil {
		return nil, err
	}

	if err := _orderChannel.GetPreloadDb(false,false, true).First(&_orderChannel,_orderChannel.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_orderChannel

	return entity, nil
}

func (CartItem) get(id uint) (Entity, error) {

	var cartItem CartItem

	err := cartItem.GetPreloadDb(false, false, true).First(&cartItem, id).Error
	if err != nil {
		return nil, err
	}
	return &cartItem, nil
}
func (cartItem *CartItem) load() error {
	if cartItem.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := cartItem.GetPreloadDb(false, true, true).First(cartItem, cartItem.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (cartItem *CartItem) loadByPublicId() error {
	return errors.New("Нет возможности найти объект по public Id")
}

func (CartItem) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return CartItem{}.getPaginationList(accountId, 0,100,sortBy,"",nil)
}

func (CartItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	orderChannels := make([]CartItem,0)
	var total uint

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&CartItem{}).GetPreloadDb(false, false, true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orderChannels, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&CartItem{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&CartItem{}).GetPreloadDb(false, false, true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orderChannels).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&CartItem{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}
	
	entities := make([]Entity,len(orderChannels))
	for i,_ := range orderChannels {
		entities[i] = &orderChannels[i]
	}

	return entities, total, nil
}

func (cartItem *CartItem) update(input map[string]interface{}) error {
	return cartItem.GetPreloadDb(false, true, false).Where("id = ?", cartItem.Id).Omit("id", "account_id").Updates(input).Error
}

func (cartItem *CartItem) delete () error {
	if err := cartItem.Amount.delete(); err != nil {
		return err
	}
	
	return db.Where("id = ?", cartItem.Id).Delete(cartItem).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############
func (cartItem *CartItem) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&cartItem)
	} else {
		_db = _db.Model(&CartItem{})
	}

	if preload {
		return _db.Preload("Amount").Preload("PaymentMethod").Preload("Product")
	} else {
		return _db
	}
}

