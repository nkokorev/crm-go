package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	"github.com/nkokorev/crm-go/utils"
)

// Прообраз торговой точки
type Shop struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not null;"`

	Name string `json:"name" gorm:"type:varchar(255);default:'Новый магазин';not null;"`
	Address string `json:"address" gorm:"type:varchar(255);default:null;"`
	Email string `json:"email" gorm:"type:varchar(255);default:null;"`
	Phone string `json:"phone" gorm:"type:varchar(255);default:null;"`

	Deliveries []Delivery  `json:"deliveries" gorm:"-"`// `gorm:"polymorphic:Owner;"`

	ProductGroups []ProductGroup `json:"productGroups"`
}

func (Shop) PgSqlCreate() {
	
	db.CreateTable(&Shop{})
	db.Model(&Shop{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Exec("ALTER TABLE shops\n    ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
}

// ############# Entity interface #############
func (shop Shop) getId() uint { return shop.ID }
func (shop *Shop) setId(id uint) { shop.ID = id }
func (shop Shop) GetAccountId() uint { return shop.AccountID }
func (shop *Shop) setAccountId(id uint) { shop.AccountID = id }
func (shop Shop) systemEntity() bool {
	return false
}
// ############# END Of Entity interface #############

func (shop *Shop) BeforeCreate(scope *gorm.Scope) error {
	shop.ID = 0
	return nil
}

func (shop *Shop) AfterFind() (err error) {

	shop.Deliveries = shop.GetDeliveryMethods()
	return nil
}

// ######### CRUD Functions ############
func (shop Shop) create() (Entity, error)  {
	var newItem Entity = &shop

	if err := db.Create(newItem).First(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (Shop) get(id uint) (Entity, error) {

	var shop Shop

	err := db.First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

func (shop *Shop) load() error {

	err := db.First(shop).Error
	if err != nil {
		return err
	}
	return nil
}


func (Shop) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	shops := make([]Shop,0)
	var total uint

	err := db.Model(&Shop{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&shops).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&Shop{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(shops))
	for i,_ := range shops {
		entities[i] = &shops[i]
	}

	return entities, total, nil
}

func (Shop) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	shops := make([]Shop,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Shop{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&shops, "name ILIKE ? OR address ILIKE ? OR email ILIKE ? OR phone ILIKE ?", search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Shop{}).
			Where("account_id = ? AND name ILIKE ? OR address ILIKE ? OR email ILIKE ? OR phone ILIKE ?", accountId, search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Shop{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&shops).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Shop{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(shops))
	for i,_ := range shops {
		entities[i] = &shops[i]
	}

	return entities, total, nil
}

func (shop *Shop) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(shop).Omit("id", "account_id").Update(input).Error
}

func (shop Shop) delete () error {
	return db.Model(Shop{}).Where("id = ?", shop.ID).Delete(shop).Error
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############

func (account Account) ExistProductGroups(groupId uint) bool {
	if groupId < 1 {
		return false
	}

	return !db.Model(&ProductGroup{}).Where("account_id = ? AND id = ?", account.ID, groupId).First(&ProductGroup{}).RecordNotFound()
}
// ######### END OF ACCOUNT Functions ############

// ######### SHOP PRODUCT Functions ############
func (shop Shop) CreateProduct(input Product, card *ProductCard) (*Product, error) {
	input.AccountID = shop.AccountID

	if input.ExistSKU() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"sku":"Товар с таким SKU уже есть"}}
	}
	if input.ExistModel() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"model":"Товар с такой моделью уже есть"}}
	}

	// Создаем продукт
	product, err := input.create()
	if err != nil {
		return nil, err
	}

	// Добавляем продукт в карточку товара
	if card != nil {
		if err = card.AppendProduct(product); err != nil {
			return nil, err
		}
	}
	
	return product, nil
}

func (shop Shop) GetProduct(productId uint) (*Product, error) {

	// Создаем продукт
	product, err := Product{}.get(productId)
	if err != nil {
		return nil, err
	}

	if product.AccountID != shop.AccountID {
		return nil, utils.Error{Message: "Продукт с указанным id не найден"}
	}

	return product, nil
}

func (shop Shop) CreateProductWithCardAndGroup(input Product, newCard ProductCard, groupId *uint) (*Product, error) {
	input.AccountID = shop.AccountID

	if input.ExistSKU() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"sku":"Товар с таким SKU уже есть"}}
	}
	if input.ExistModel() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"model":"Товар с такой моделью уже есть"}}
	}

	// Создаем продукт
	product, err := input.create()
	if err != nil {
		return nil, err
	}

	// Создаем карточку товара
	newCard.AccountID = shop.AccountID
	newCard.ShopID = shop.ID
	if groupId != nil {
		newCard.ProductGroupID = groupId
	}
	card, err := newCard.create()
	if err != nil {
		return product, err
	}

	// Добавляем товар в новую карточку
	if err = card.AppendProduct(product); err != nil {
		return nil, err
	}

	return product, nil
}

/////////////////////////

func (shop Shop) AppendDeliveryMethod(entity Entity) error {
	return entity.update(map[string]interface{}{"shop_id":shop.ID})
}

func (shop Shop) GetDeliveryMethods() []Delivery {
	// Находим все необходимые методы
	var posts []DeliveryRussianPost
	if err := db.Model(&DeliveryRussianPost{}).Find(&posts, "account_id = ? AND shop_id = ?", shop.AccountID, shop.ID).Error; err != nil {
		return nil
	}

	var couriers []DeliveryCourier
	if err := db.Model(&DeliveryCourier{}).Find(&couriers, "account_id = ? AND shop_id = ?", shop.AccountID, shop.ID).Error; err != nil {
		return nil
	}

	var pickups []DeliveryPickup
	if err := db.Model(&DeliveryPickup{}).Find(&pickups, "account_id = ? AND shop_id = ?", shop.AccountID, shop.ID).Error; err != nil {
		return nil
	}


	deliveries := make([]Delivery, len(posts)+len(pickups)+len(couriers))
	for i,_ := range posts {
		deliveries[i] = &posts[i]
	}
	for i,_ := range couriers {
		deliveries[i+len(posts)] = &couriers[i]
	}
	for i,_ := range pickups {
		deliveries[i+len(posts)+len(couriers)] = &pickups[i]
	}

/*	fmt.Println("New list: ")
	for i,v := range deliveries {
		fmt.Printf("[%s], %v\n\r",i, v)
	}*/

	return deliveries
}

func (shop Shop) CalculateDelivery(deliveryRequest DeliveryRequest) (*DeliveryData, error) {

	delivery, err := shop.GetDelivery(deliveryRequest.DeliveryMethod.Code, deliveryRequest.DeliveryMethod.ID)
	if err != nil {
		return nil, err
	}

	// 1. Расчет веса посылки
	if deliveryRequest.DeliveryData.NeedToCalculateWeight {
		var weight float64
		weight = 0

		// проходим циклом по продуктам и складываем их общий вес
		for _,v := range deliveryRequest.Cart {
			// 1. Получаем продукт
			product, err := shop.GetProduct(v.ProductId)
			if err != nil {
				return nil, err
			}

			_w, err := product.GetAttribute(deliveryRequest.DeliveryData.ProductWeightKey)
			wg, ok := _w.(float64)
			if !ok {
				continue
			}
			weight += wg * float64(v.Count)
		}

		deliveryRequest.DeliveryData.Weight = weight
	}

	// 2. Проверяем максимальный вес
	if err := delivery.checkMaxWeight(deliveryRequest.DeliveryData); err != nil {
		fmt.Println("Ошибка макс веса", err)
		return nil, err
	}

	// 3. Проводим расчет стоимости доставки
	deliveryData, err := delivery.CalculateDelivery(deliveryRequest.DeliveryData)
	if err != nil {
		return nil, err
	}

	return deliveryData, nil
}

func (shop Shop) CreateDelivery(input map[string]interface{}) (*Entity, error) {

	var delivery Delivery

	codeStr, ok := input["code"];
	if !ok {
		return nil, utils.Error{Message: "Не хватает параметра code в запросе"}
	}

	code, ok := codeStr.(string)

	switch code {
	case "russianPost":
		var deliveryRussianPost DeliveryRussianPost
		if err := mapstructure.Decode(input, &deliveryRussianPost); err != nil {
			return nil, err
		}
		delivery = &deliveryRussianPost
	case "courier":
		var deliveryCourier DeliveryCourier
		if err := mapstructure.Decode(input, &deliveryCourier); err != nil {
			return nil, err
		}
		delivery = &deliveryCourier
	case "pickup":
		var deliveryPickup DeliveryPickup
		if err := mapstructure.Decode(input, &deliveryPickup); err != nil {
			return nil, err
		}
		delivery = &deliveryPickup
	default:
		return nil, utils.Error{Message: "Ошибка в коде типа создаваемого интерфейса"}
	}

	delivery.setShopId(shop.ID)
	delivery.setAccountId(shop.AccountID)

	entity, err := delivery.create()
	if err != nil {
		return nil, err
	}
	return &entity, nil

	// return &delivery, nil
}

func (shop Shop) GetDelivery(code string, methodId uint) (Delivery, error) {

	// Получаем все варианты доставки (обычно их мало). Можно через switch, но лень потом исправлять баг с новыми типом доставки
	deliveries := shop.GetDeliveryMethods()

	// Ищем наш вариант доставки
	var delivery Delivery
	for _,v := range deliveries {
		if v.GetCode() == code && v.getId() == methodId {
			delivery = v
			break
		}
	}

	// Проверяем, удалось ли найти выбранный вариант доставки
	if delivery == nil {
		return nil, utils.Error{Message: "Не верно указан тип доставки"}
	}

	return delivery, nil
}
// 
func (shop Shop) UpdateDelivery(input map[string]interface{}) (Delivery, error) {

	// Парсим тип рассылки и ее ID
	code,ok := input["code"].(string)
	methodId, ok := input["id"].(float64)
	if !ok {
		return nil, utils.Error{Message: "Код или id способа доставки не верен"}
	}

	delivery, err := shop.GetDelivery(code, uint(methodId))
	if err != nil {
		return nil, err
	}

	if delivery.GetAccountId() != shop.AccountID {
		return nil, utils.Error{Message: "Метод доставки принадлежит другому аккаунту!"}
	}

	err = delivery.update(input)
	if err != nil {
		return nil, err
	}

	return delivery, nil
}

func (shop Shop) DeleteDelivery(input map[string]interface{}) error {

	// Парсим тип рассылки и ее ID
	code,ok := input["code"].(string)
	methodId, ok := input["id"].(float64)
	if !ok {
		return utils.Error{Message: "Код или id способа доставки не верен"}
	}

	delivery, err := shop.GetDelivery(code, uint(methodId))
	if err != nil {
		return err
	}

	if delivery.GetAccountId() != shop.AccountID {
		return utils.Error{Message: "Метод доставки принадлежит другому аккаунту!"}
	}

	err = delivery.delete()
	if err != nil {
		return err
	}

	return nil
}

func (shop Shop) DeliveryListOptions() map[string]interface{} {
	return map[string]interface{}{
		"russianPost": "Почта России",
		"courier": "Курьерская доставка",
		"pickup": "Самовывоз",
	}
}

// Вспомогательная функция
func GetAccountIdByShopId(shopId uint) (uint, error) {
	type Result struct {
		AccountId uint `json:"accountId"`
	}
	var result Result
	if err := db.Table("shops").Where("shop_id = ?", shopId).Scan(&result).Error; err != nil {
		return 0, err
	}

	return result.AccountId, nil
}
/*,*/