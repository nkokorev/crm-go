package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

// Корзины товаров
type CartItem struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа
	OrderId 	uint	`json:"orderId" gorm:"index;not null"` // заказ, к которому относится корзина

	ProductId	uint    // позиции товаров
	Number		uint	// число товаров

	Product 	Product `json:"product" gorm:"preload:false"`
	Order	 	Product `json:"product" gorm:"preload:false"`

	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (orderComment CartItem) GetId() uint { return orderComment.Id }
func (orderComment *CartItem) setId(id uint) { orderComment.Id = id }
func (orderComment CartItem) GetAccountId() uint { return orderComment.AccountId }
func (orderComment *CartItem) setAccountId(id uint) { orderComment.AccountId = id }
func (orderComment CartItem) SystemEntity() bool { return orderComment.AccountId == 1 }

// ############# Entity interface #############

func (CartItem) PgSqlCreate() {
	db.CreateTable(&CartItem{})
	db.Model(&CartItem{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}
func (orderComment *CartItem) BeforeCreate(scope *gorm.Scope) error {
	orderComment.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (orderComment CartItem) create() (Entity, error)  {
	_orderChannel := orderComment
	if err := db.Create(&_orderChannel).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_orderChannel

	return entity, nil
}

func (CartItem) get(id uint) (Entity, error) {

	var orderComment CartItem

	err := db.First(&orderComment, id).Error
	if err != nil {
		return nil, err
	}
	return &orderComment, nil
}
func (orderComment *CartItem) load() error {
	if orderComment.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := db.First(orderComment, orderComment.Id).Error
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

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&CartItem{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

		err := db.Model(&CartItem{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

	// Преобразуем полученные данные
	entities := make([]Entity,len(orderChannels))
	for i,_ := range orderChannels {
		entities[i] = &orderChannels[i]
	}

	return entities, total, nil
}

func (orderComment *CartItem) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(orderComment).Omit("id", "account_id").Updates(input).Error
}

func (orderComment CartItem) delete () error {
	return db.Model(CartItem{}).Where("id = ?", orderComment.Id).Delete(orderComment).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

