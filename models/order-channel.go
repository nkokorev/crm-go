package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"time"
)

type OrderChannel struct {
	Id     uint   `json:"id" gorm:"primary_key"`
	AccountId uint `json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа

	Name		string 	`json:"name" gorm:"type:varchar(255);not null;"` // "Заказ из корзины", "Заказ по телефону", "Пропущенный звонок", "Письмо.."
	Description string 	`json:"description" gorm:"type:varchar(255);"` // Описание назначения канала
	
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (orderChannel OrderChannel) GetId() uint { return orderChannel.Id }
func (orderChannel *OrderChannel) setId(id uint) { orderChannel.Id = id }
func (orderChannel OrderChannel) GetAccountId() uint { return orderChannel.AccountId }
func (orderChannel *OrderChannel) setAccountId(id uint) { orderChannel.AccountId = id }
func (orderChannel OrderChannel) SystemEntity() bool { return orderChannel.AccountId == 1 }

// ############# Entity interface #############

func (OrderChannel) PgSqlCreate() {
	if !db.HasTable(&OrderChannel{}) {
		db.CreateTable(&OrderChannel{})
	}
	db.Model(&OrderChannel{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	orderChannels := []OrderChannel {
		{Name:   "Оффлайн",		Description: "-"},
		{Name:   "По телефону",	Description: "-"},
		{Name:   "Пропущенный звонок",	Description: "-"},
		{Name:   "Через корзину",	Description: "-"},
		{Name:   "В один клик",		Description: "Экспресс заказ"},
		{Name:   "Запрос на понижение цены",	Description: "-"},
		{Name:   "Заявка с посадочной страницы",Description: "-"},
		{Name:   "Мессенджеры",	Description: "-"},
		{Name:   "Онлайн-консультант",Description: "-"},
		{Name:   "Мобильное приложение",Description: "-"},
	}

	for i := range(orderChannels) {
		_, err := Account{Id: 1}.CreateEntity(&orderChannels[i])
		if err != nil {
			log.Fatalf("Не удалось создать orderChannels: ", err)
		}
	}
	
}
func (orderChannel *OrderChannel) BeforeCreate(scope *gorm.Scope) error {
	orderChannel.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (orderChannel OrderChannel) create() (Entity, error)  {
	_orderChannel := orderChannel
	if err := db.Create(&_orderChannel).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_orderChannel

	return entity, nil
}

func (OrderChannel) get(id uint) (Entity, error) {

	var orderChannel OrderChannel

	err := db.First(&orderChannel, id).Error
	if err != nil {
		return nil, err
	}
	return &orderChannel, nil
}
func (orderChannel *OrderChannel) load() error {
	if orderChannel.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить OrderChannel - не указан  Id"}
	}

	err := db.First(orderChannel, orderChannel.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (OrderChannel) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return OrderChannel{}.getPaginationList(accountId, 0,100,sortBy,"")
}

func (OrderChannel) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	orderChannels := make([]OrderChannel,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&OrderChannel{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&orderChannels, "name ILIKE ? OR description ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&OrderChannel{}).
			Where("account_id IN (?) AND name ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&OrderChannel{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&orderChannels).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&OrderChannel{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
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

func (orderChannel *OrderChannel) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(orderChannel).Where("id", orderChannel.Id).Omit("id", "account_id").Updates(input).Error
}

func (orderChannel OrderChannel) delete () error {
	return db.Model(OrderChannel{}).Where("id = ?", orderChannel.Id).Delete(orderChannel).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############
