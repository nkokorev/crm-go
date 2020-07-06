package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
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

	Delivery []Delivery  `gorm:"-"`// `gorm:"polymorphic:Owner;"`

	ProductGroups []ProductGroup `json:"productGroups"`
}

func (Shop) PgSqlCreate() {
	
	db.CreateTable(&Shop{})
	db.Exec("ALTER TABLE shops\n    ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
}

func (shop *Shop) BeforeCreate(scope *gorm.Scope) error {
	shop.ID = 0
	return nil
}

func (shop Shop) getId() uint {
	return shop.ID
}

// ######### CRUD Functions ############
func (shop Shop) create() (*Shop, error)  {
	var shopNew = shop
	err := db.Create(&shopNew).Error
	return &shopNew, err
}

func (Shop) get(id uint) (*Shop, error) {

	shop := Shop{}

	if err := db.Table("shops").Preload("ProductGroups").First(&shop, id).Error; err != nil {
		return nil, err
	}

	return &shop, nil
}

func (Shop) getList(accountId uint) ([]Shop, error) {

	shops := make([]Shop,0)

	err := db.Find(&shops, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return shops, nil
}

func (shop *Shop) update(input interface{}) error {
	return db.Model(shop).Select("name", "address").Where("id = ?", shop.ID).
		Updates(structs.Map(input)).Error
}

func (shop Shop) delete () error {
	return db.Model(Shop{}).Where("id = ?", shop.ID).Delete(shop).Error
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############
func (account Account) CreateShop(input Shop) (*Shop, error) {
	input.AccountID = account.ID
	shop, err := input.create()
	if err != nil {
		return nil, err
	}

	go account.CallWebHookIfExist(EventShopCreated, shop)

	return shop, nil
}

func (account Account) GetShop(id uint) (*Shop, error) {
	shop, err := Shop{}.get(id)
	if err != nil {
		return nil, err
	}

	if account.ID != shop.AccountID {
		return nil, utils.Error{Message: "Магазин принадлежит другому аккаунту"}
	}

	return shop, nil
}

func (account Account) GetShops() ([]Shop, error) {
	return Shop{}.getList(account.ID)
}

func (account Account) UpdateShop(id uint, input interface{}) (*Shop, error) {
	shop, err := account.GetShop(id)
	if err != nil {
		return nil, err
	}

	if account.ID != shop.AccountID {
		return nil, utils.Error{Message: "Магазин принадлежит другому аккаунту"}
	}

	err = shop.update(input)

	return shop, err

}

func (account Account) DeleteShop(id uint) error {

	// включает в себя проверку принадлежности к аккаунту
	shop, err := account.GetShop(id)
	if err != nil {
		return err
	}

	err = shop.delete()
	if err != nil {
		return err;
	}

	go account.CallWebHookIfExist(EventShopDeleted, shop)

	return nil
}

func (account Account) ExistShop(id uint) bool {
	if id < 1 {
		return false
	}
	return !db.Model(&Shop{}).Where("account_id = ? AND id = ?", account.ID, id).First(&Shop{}).RecordNotFound()
}

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
/*func (shop Shop) AppendDeliveryMethod(deliveryPostRussia DeliveryRussianPost) error {
	return deliveryPostRussia.update(map[string]interface{}{"shop_id":shop.ID})
	// return db.Model(&shop).Association("Delivers").Append(deliveryPostRussia).Error
}*/

/*func (deliveryRussianPost DeliveryRussianPost) AppendAssociationMethod(deliveryOption DeliveryOption) error {
	fmt.Println("deliveryRussianPost: ", deliveryRussianPost)
	fmt.Println("deliveryOption: ", deliveryOption)
	return db.Model(&deliveryRussianPost).Association("Delivers").Append(deliveryOption).Error
}*/

