package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// Внутренние комментарии менеджеров / ответственных
type OrderComment struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа
	UserId 		uint	`json:"userId" gorm:"index,not null"`

	Description string 	`json:"description" gorm:"type:varchar(255);"` // Описание назначения канала
	ManagersComments string `json:"description" gorm:"type:varchar(255);"`
	
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (orderComment OrderComment) GetId() uint { return orderComment.Id }
func (orderComment *OrderComment) setId(id uint) { orderComment.Id = id }
func (orderComment OrderComment) GetAccountId() uint { return orderComment.AccountId }
func (orderComment *OrderComment) setAccountId(id uint) { orderComment.AccountId = id }
func (orderComment OrderComment) SystemEntity() bool { return orderComment.AccountId == 1 }

// ############# Entity interface #############

func (OrderComment) PgSqlCreate() {
	db.CreateTable(&OrderComment{})
	db.Model(&OrderComment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}
func (orderComment *OrderComment) BeforeCreate(scope *gorm.Scope) error {
	orderComment.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (orderComment OrderComment) create() (Entity, error)  {
	_orderChannel := orderComment
	if err := db.Create(&_orderChannel).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_orderChannel

	return entity, nil
}

func (OrderComment) get(id uint) (Entity, error) {

	var orderComment OrderComment

	err := db.First(&orderComment, id).Error
	if err != nil {
		return nil, err
	}
	return &orderComment, nil
}
func (orderComment *OrderComment) load() error {
	if orderComment.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить OrderComment - не указан  Id"}
	}

	err := db.First(orderComment, orderComment.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (OrderComment) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	orderChannels := make([]OrderComment,0)
	var total uint

	err := db.Model(&OrderComment{}).Limit(100).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&orderChannels).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&OrderComment{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(orderChannels))
	for i,_ := range orderChannels {
		entities[i] = &orderChannels[i]
	}

	return entities, total, nil
}

func (OrderComment) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	orderChannels := make([]OrderComment,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&OrderComment{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orderChannels, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&OrderComment{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&OrderComment{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orderChannels).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&OrderComment{}).Where("account_id = ?", accountId).Count(&total).Error
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

func (orderComment *OrderComment) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(orderComment).Omit("id", "account_id").Updates(input).Error
}

func (orderComment OrderComment) delete () error {
	return db.Model(OrderComment{}).Where("id = ?", orderComment.Id).Delete(orderComment).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

