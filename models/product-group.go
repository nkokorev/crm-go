package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

type ProductGroup struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	ShopID uint `json:"shopId" gorm:"type:int;index;not_null;"` // магазин, к которому относится данная группа
	// AccountID uint `json:"-" gorm:"type:int;index;not_null;"` // хз хз
	ParentID uint `json:"parentId" gorm:"default:NULL"`
	
	Code string `json:"code" gorm:"default:NULL"` // tea, coffe, china
	URL string `json:"url" gorm:"default:NULL"`

	Name string `json:"name" gorm:"default:NULL"` // Чай, кофе, ..

	Breadcrumb string `json:"breadcrumb" gorm:"default:NULL"`
	ShortDescription string `json:"shortDescription" gorm:"default:NULL"`
	Description string `json:"description" gorm:"default:NULL"`

	MetaTitle string `json:"metaTitle"`
	MetaKeywords string `json:"metaKeywords"`
	MetaDescription string `json:"metaDescription"`

	Shop Shop `json:"shop" `
	ParentGroup *ProductGroup `json:"-"`
	Products []Product `json:"products" gorm:"many2many:product_group_products"`

}

func (ProductGroup) PgSqlCreate() {
	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&ProductGroup{})
	db.Exec("ALTER TABLE product_groups\n--     ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT products_shop_id_fkey FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT products_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
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
	err := db.Create(group).Error
	return &group, err
}

func (ProductGroup) get(id uint) (*ProductGroup, error) {

	group := ProductGroup{}

	if err := db.Table(ProductGroup{}.TableName()).Preload("Shop").First(&group, id).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (ProductGroup) getList(shopId uint) ([]ProductGroup, error) {

	groups := make([]ProductGroup,0)

	err := db.Find(&groups, "shop_id = ?", shopId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return groups, nil
}

func (group *ProductGroup) update(input interface{}) error {
	return db.Model(group).Select("Name", "Address").Updates(structs.Map(input)).Error

}

func (group ProductGroup) delete () error {
	return db.Model(Shop{}).Where("id = ?", group.ID).Delete(group).Error
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
	if group.ShopID != shop.ID {
		return nil, utils.Error{Message: "Группа принадлежит другому магазину"}
	}

	return group, nil
}

func (shop Shop) GetProductGroups() ([]ProductGroup, error) {
	return ProductGroup{}.getList(shop.ID)
}

func (shop Shop) UpdateProductGroup(groupId uint, input interface{}) (*ProductGroup, error) {
	group, err := shop.GetProductGroup(groupId)
	if err != nil {
		return nil, err
	}

	if group.ShopID != shop.ID {
		return nil, utils.Error{Message: "Группа принадлежит другому магазину"}
	}

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

// Создает и добавляет продукт в продуктовую группу

func (group ProductGroup) CreateProduct(product Product) (*Product, error) {
	// add account id
	// add group id
	return nil, nil
}

func (group ProductGroup) GetProductList() ([]Product, error) {

	products := make([]Product,0)

	err := db.Model(&group).Association("Products").Find(&products).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return products, nil
}

// ######### END OF ProductGroup Functions ############




