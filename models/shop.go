package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

// Прообраз торговой точки
type Shop struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not_null;"`

	Name string `json:"name" gorm:"type:varchar(255);default:'Новый магазин';not_null;"`
	Address string `json:"address" gorm:"type:varchar(255);default:null;"`
}

func (Shop) PgSqlCreate() {
	
	db.CreateTable(&Shop{})
	db.Exec("ALTER TABLE shops\n    ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
}

func (shop *Shop) BeforeCreate(scope *gorm.Scope) error {
	shop.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (Shop) create(input Shop) (*Shop, error)  {
	var shop = input
	err := db.Create(shop).Error
	return &shop, err
}

func (Shop) get(id uint) (*Shop, error) {

	shop := Shop{}

	if err := db.First(&shop, id).Error; err != nil {
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
	return db.Model(shop).Select("Name", "Address").Updates(structs.Map(input)).Error

}

func (shop Shop) delete () error {
	return db.Model(Shop{}).Where("id = ?", shop.ID).Delete(shop).Error
}


// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############

func (account Account) CreateShop(input Shop) (*Shop, error) {
	input.AccountID = account.ID
	return Shop{}.create(input)
}

func (account Account) GetShop(productId uint) (*Shop, error) {
	shop, err := Shop{}.get(productId)
	if err != nil {
		return nil, err
	}

	if account.ID != shop.AccountID {
		return nil, utils.Error{Message: "Магазин принадлежит другому аккаунту"}
	}

	return nil, nil
}

func (account Account) GetShops() ([]Shop, error) {
	return Shop{}.getList(account.ID)
}

func (account Account) UpdateShop(productId uint, input interface{}) (*Shop, error) {
	shop, err := account.GetShop(productId)
	if err != nil {
		return nil, err
	}

	if account.ID != shop.AccountID {
		return nil, utils.Error{Message: "Магазин принадлежит другому аккаунту"}
	}

	err = shop.update(input)

	return shop, err

}

func (account Account) DeleteShop(productId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	shop, err := account.GetShop(productId)
	if err != nil {
		return err
	}

	return shop.delete()
}


// ######### END OF ACCOUNT Functions ############

func (shop Shop) CreateProductGroup(group ProductGroup) (*ProductGroup, error) {
	return nil, nil
}