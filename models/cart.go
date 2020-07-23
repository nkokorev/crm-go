package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// Корзины товаров
type Cart struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа

	ProductId	uint    // позиции товаров
	Number		uint	// число товаров


	Products 		[]Product `json:"products" gorm:"many2many:orders_products;preload"`

	Description string 	`json:"description" gorm:"type:varchar(255);"` // Описание назначения канала
	ManagersComments string `json:"description" gorm:"type:varchar(255);"`
	
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (orderComment Cart) GetId() uint { return orderComment.Id }
func (orderComment *Cart) setId(id uint) { orderComment.Id = id }
func (orderComment Cart) GetAccountId() uint { return orderComment.AccountId }
func (orderComment *Cart) setAccountId(id uint) { orderComment.AccountId = id }
func (orderComment Cart) SystemEntity() bool { return orderComment.AccountId == 1 }

// ############# Entity interface #############

func (Cart) PgSqlCreate() {
	db.CreateTable(&Cart{})
	db.Model(&Cart{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}
func (orderComment *Cart) BeforeCreate(scope *gorm.Scope) error {
	orderComment.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (orderComment Cart) create() (Entity, error)  {
	_orderChannel := orderComment
	if err := db.Create(&_orderChannel).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_orderChannel

	return entity, nil
}

func (Cart) get(id uint) (Entity, error) {

	var orderComment Cart

	err := db.First(&orderComment, id).Error
	if err != nil {
		return nil, err
	}
	return &orderComment, nil
}
func (orderComment *Cart) load() error {
	if orderComment.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Cart - не указан  Id"}
	}

	err := db.First(orderComment, orderComment.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (Cart) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	return Cart{}.getPaginationList(accountId, 0,100,sortBy,"")
}

func (Cart) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	orderChannels := make([]Cart,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Cart{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orderChannels, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Cart{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Cart{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orderChannels).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Cart{}).Where("account_id = ?", accountId).Count(&total).Error
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

func (orderComment *Cart) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(orderComment).Omit("id", "account_id").Updates(input).Error
}

func (orderComment Cart) delete () error {
	return db.Model(Cart{}).Where("id = ?", orderComment.Id).Delete(orderComment).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

