package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type OrderChannel struct {
	Id     		uint    `json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"account_id" gorm:"index;not null"` // аккаунт-владелец ключа

	Code		string	`json:"code" gorm:"type:varchar(32);unique;not null;"`
	Name		string 	`json:"name" gorm:"type:varchar(255);not null;"` // "Заказ из корзины", "Заказ по телефону", "Пропущенный звонок", "Письмо.."
	Description string 	`json:"description" gorm:"type:varchar(255);"` // Описание назначения канала
}

// ############# Entity interface #############
func (orderChannel OrderChannel) GetId() uint { return orderChannel.Id }
func (orderChannel *OrderChannel) setId(id uint) { orderChannel.Id = id }
func (orderChannel *OrderChannel) setPublicId(id uint) { }
func (orderChannel OrderChannel) GetAccountId() uint { return orderChannel.AccountId }
func (orderChannel *OrderChannel) setAccountId(id uint) { orderChannel.AccountId = id }
func (orderChannel OrderChannel) SystemEntity() bool { return orderChannel.AccountId == 1 }

// ############# Entity interface #############

func (OrderChannel) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&OrderChannel{}); err != nil {log.Fatal(err)}
	// db.Model(&OrderChannel{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE order_channels ADD CONSTRAINT order_channels_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
	
	db.Delete(&OrderChannel{},"id > 0")
	orderChannels := []OrderChannel {
		{Code:"offline", Name:   "Оффлайн",		Description: "-"},
		{Code:"phone", 			Name:   "По телефону",	Description: "-"},
		{Code:"missed_call", 	Name:   "Пропущенный звонок",	Description: "-"},
		{Code:"through_the_basket", Name:   "Через корзину",	Description: "-"},
		{Code:"with_one_click", 	Name:   "В один клик",		Description: "Экспресс заказ"},
		{Code:"request_to_lower_the_price", 	Name:   "Запрос на понижение цены",	Description: "-"},
		{Code:"request_from_the_landing_page",	Name:   "Заявка с посадочной страницы",Description: "-"},
		{Code:"messenger", 			Name:   "Мессенджеры",	Description: "-"},
		{Code:"online_assistant", 	Name:   "Онлайн-консультант",Description: "-"},
		{Code:"mobile_apps", 	Name:   "Мобильное приложение",Description: "-"},
		{Code:"callback_phone", Name:   "Заказ обратного звонка",Description: "-"},
		{Code:"callback_form", Name:   "Вопрос с формы на сайте",Description: "-"},
	}
	for i := range(orderChannels) {
		_, err := Account{Id: 1}.CreateEntity(&orderChannels[i])
		if err != nil {
			log.Fatalf("Не удалось создать orderChannels: ", err)
		}
	}
	
}
func (orderChannel *OrderChannel) BeforeCreate(tx *gorm.DB) error {
	orderChannel.Id = 0
	return nil
}
func (orderChannel *OrderChannel) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(orderChannel)
	} else {
		_db = _db.Model(&OrderChannel{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{""})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ######### CRUD Functions ############
func (orderChannel OrderChannel) create() (Entity, error)  {
	_item := orderChannel
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (OrderChannel) get(id uint, preloads []string) (Entity, error) {

	var item OrderChannel

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (orderChannel *OrderChannel) load(preloads []string) error {
	if orderChannel.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить OrderChannel - не указан  Id"}
	}

	err := orderChannel.GetPreloadDb(false, false, preloads).First(orderChannel, orderChannel.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*OrderChannel) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
// Специальная функция *** NOT Entity ***
func (OrderChannel) getByCode(accountId uint, code string) (*OrderChannel, error) {
	var orderChannel OrderChannel
	err := db.First(&orderChannel,"account_id IN (?) AND code = ?", []uint{1, accountId}, code).Error
	return &orderChannel, err
}

func (OrderChannel) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return OrderChannel{}.getPaginationList(accountId, 0,100,sortBy,"",nil,preload)
}
func (OrderChannel) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	orderChannels := make([]OrderChannel,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&OrderChannel{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&orderChannels, "name ILIKE ? OR description ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&OrderChannel{}).GetPreloadDb(false, false, nil).
			Where("account_id IN (?) AND name ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&OrderChannel{}).GetPreloadDb(false, false, preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&orderChannels).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&OrderChannel{}).GetPreloadDb(false, false, nil).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(orderChannels))
	for i := range orderChannels {
		entities[i] = &orderChannels[i]
	}

	return entities, total, nil
}

func (orderChannel *OrderChannel) update(input map[string]interface{}, preloads []string) error {

	// delete(input,"amount")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/

	if err := orderChannel.GetPreloadDb(false, false, nil).Where("id = ?", orderChannel.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := orderChannel.GetPreloadDb(false,false, preloads).First(orderChannel, orderChannel.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (orderChannel *OrderChannel) delete () error {
	return orderChannel.GetPreloadDb(true,false,nil).Where("id = ?", orderChannel.Id).Delete(orderChannel).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

func (account Account) GetOrderChannelByCode(code string) (*OrderChannel, error){
	return OrderChannel{}.getByCode(account.Id, code)
}
