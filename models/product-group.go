package models

import (
	"github.com/jinzhu/gorm"
)

type ProductGroup struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	ShopID uint `json:"shopId" gorm:"type:int;index;not_null;"` // магазин, к которому относится данная группа
	// AccountID uint `json:"-" gorm:"type:int;index;not_null;"` // хз хз
	// ParentID uint `json:"parentId,omitempty" gorm:"default:NULL"`
	ParentID *uint `json:"parentId" gorm:"default:NULL"`

	Code string `json:"code" gorm:"type:varchar(255);default:null;"` // tea, coffe, china
	URL string `json:"url"gorm:"type:varchar(255);default:null;"`

	Name string `json:"name" gorm:"type:varchar(255);default:null;"` // Чай, кофе, ..

	Breadcrumb string `json:"breadcrumb" gorm:"type:varchar(255);default:null;"`
	ShortDescription string `json:"shortDescription" gorm:"type:varchar(255);default:null;"`
	Description string `json:"description" gorm:"type:text;default:null;"`

	MetaTitle string `json:"metaTitle" gorm:"type:varchar(255);default:null;"`
	MetaKeywords string `json:"metaKeywords" gorm:"type:varchar(255);default:null;"`
	MetaDescription string `json:"metaDescription" gorm:"type:varchar(255);default:null;"`

	Shop Shop `json:"shop" `
	ParentGroup *ProductGroup `json:"-"` // if has parentId
	// ProductCards ProductCard `json:"productCards" gorm:"many2many:product_group_product_cards"`
	ProductCards []ProductCard `json:"productCards"`
	// Products []Product `json:"products" gorm:"many2many:product_group_products"`

}

func (ProductGroup) PgSqlCreate() {
	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&ProductGroup{})
	db.Exec("ALTER TABLE product_groups\n--     ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT products_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    alter column parent_id SET DEFAULT NULL;\n--     ADD CONSTRAINT products_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\n\n-- create unique index uix_products_account_id_sku ON products (account_id,sku);\n-- alter table product_groups alter column parent_id set default NULL;\n")
}

func (group *ProductGroup) BeforeCreate(scope *gorm.Scope) error {
	group.ID = 0
	return nil
}

func (ProductGroup) TableName() string {
	return "product_groups"
}

// ######### CRUD Functions ############
func (input ProductGroup) create() (*ProductGroup, error)  {
	var group = input
	err := db.Create(&group).Error
	return &group, err
}

func (ProductGroup) get(id uint) (*ProductGroup, error) {

	group := ProductGroup{}

	if err := db.First(&group, id).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (group *ProductGroup) update(input interface{}) error {
	// return db.Model(group).Omit("id").Updates(structs.Map(input)).Error
	return db.Model(group).Omit("id").Updates(input).Error

}

func (group ProductGroup) delete () error {
	return db.Model(ProductGroup{}).Where("id = ?", group.ID).Delete(group).Error
}
// ######### END CRUD Functions ############


// ######### SHOP Functions ############
func (shop Shop) CreateProductGroup(input ProductGroup) (*ProductGroup, error) {
	input.ShopID = shop.ID
	return input.create()
}

func (shop Shop) GetProductGroup(groupId uint) (*ProductGroup, error) {
	group, err := ProductGroup{}.get(groupId)
	if err != nil {
		return nil, err
	}

	// Делаем проверку на всякий пожарный
	/*if group.ShopID != shop.ID {
		return nil, utils.Error{Message: "Группа принадлежит другому магазину"}
	}*/
	/*if group.ShopID != shop.ID {
		acc, err := GetAccount(shop.AccountID)
		if err != nil {
			return nil, utils.Error{Message: "Не удалось получить аккаунт"}
		}
		if !acc.ExistShop(group.ShopID) {
			return nil, utils.Error{Message: "Указанный магазин принадлежит другому аккаунту"}
		}
	}*/

	return group, nil
}

func (shop Shop) GetProductGroups() ([]ProductGroup, error) {

	groups := make([]ProductGroup,0)

	if err := db.Model(&shop).Association("ProductGroups").Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

func (shop Shop) UpdateProductGroup(groupId uint, input interface{}) (*ProductGroup, error) {
	group, err := shop.GetProductGroup(groupId)
	if err != nil {
		return nil, err
	}

	// Проверим, что новый Shop принадлежит этому аккаунту
	/*m := structs.Map(input)

	// todo: 
	if m["ShopID"] != shop.ID {
	   acc, err := GetAccount(shop.AccountID)
	   if err != nil {
		   return nil, utils.Error{Message: "Не удалось получить аккаунт"}
	   }
		if !acc.ExistShop(group.ShopID) {
			return nil, utils.Error{Message: "Указанный магазин принадлежит другому аккаунту"}
		}
	}

	// если меняется родитель
	if m["ParentID"] != group.ParentID {
		// todo: 
	}*/

	err = group.update(input)
	if err != nil {
		return nil, err
	}

	return group, nil

}

func (shop Shop) DeleteProductGroup(groupId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	group, err := shop.GetProductGroup(groupId)
	if err != nil {
		return err
	}

	return group.delete()
}

// ######### END OF SHOP Functions ############


// ######### ProductGroup Functions ############
func (pg ProductGroup) CreateChild(input ProductGroup) (*ProductGroup, error) {
	input.ParentID = &pg.ID
	input.ShopID = pg.ShopID
	return input.create()
}

// Создает и добавляет продукт в категорию товаров
func (group ProductGroup) CreateProductCard(_c *ProductCard) (*ProductCard, error) {
 	_c.ProductGroupID = group.ID
	return _c.create()
}
/*func (group ProductGroup) CreateAndAppendProductCard(card *ProductCard) error {
	return db.Model(&group).Association("ProductCards").Append(card).Error
}*/

func (group ProductGroup) GetProductCards() ([]ProductCard, error) {

	cards := make([]ProductCard,0)

	if err := db.Model(&group).Association("ProductCards").Find(&cards).Error; err != nil {
		return nil, err
	}

	return cards, nil
}

// ######### END OF ProductGroup Functions ############




