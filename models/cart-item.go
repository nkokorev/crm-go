package models

import (
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

	// Ставка НДС
	VatCode	uint	`json:"vat_code"`

	Product Product `json:"product"`
	Order	Order `json:"order" gorm:"preload:false"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (CartItem) PgSqlCreate() {
	db.CreateTable(&CartItem{})
	db.Model(&CartItem{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&CartItem{}).AddForeignKey("amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
	db.Model(&CartItem{}).AddForeignKey("order_id", "orders(id)", "CASCADE", "CASCADE")

}
func (cartItem *CartItem) BeforeCreate(scope *gorm.Scope) error {
	cartItem.Id = 0
	cartItem.Amount.AccountId = cartItem.AccountId
	return nil
}

// ############# Entity interface #############
func (cartItem CartItem) GetId() uint { return cartItem.Id }
func (cartItem *CartItem) setId(id uint) { cartItem.Id = id }
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

	var entity Entity = &_orderChannel

	return entity, nil
}

func (CartItem) get(id uint) (Entity, error) {

	var cartItem CartItem

	err := db.Preload("Amount").Preload("Product").Preload("Product.PaymentSubject").First(&cartItem, id).Error
	if err != nil {
		return nil, err
	}
	return &cartItem, nil
}
func (cartItem *CartItem) load() error {
	if cartItem.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := db.Preload("Amount").Preload("Product").Preload("Product.PaymentSubject").First(cartItem, cartItem.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (CartItem) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return CartItem{}.getPaginationList(accountId, 0,100,sortBy,"")
}

func (CartItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	orderChannels := make([]CartItem,0)
	var total uint

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&CartItem{}).Preload("Amount").Preload("Product").Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

		err := db.Model(&CartItem{}).Preload("Amount").Preload("Product").Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
	return db.Set("gorm:association_autoupdate", false).
		Model(cartItem).Omit("id", "account_id").Updates(input).Error
}

func (cartItem *CartItem) delete () error {
	if err := cartItem.Amount.delete(); err != nil {
		return err
	}
	
	return db.Where("id = ?", cartItem.Id).Delete(cartItem).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

