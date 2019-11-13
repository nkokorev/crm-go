package models

// тип атрибута
type EavAttr struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	AccountID 	uint 	`json:"-" gorm:"default:null;"` // если роль системная, то нет accountID
	Name 		string 	`json:"name" gorm:"not null;"` // "Тип сырья", "Цвет", "Размер", "Вес"
	System 		bool 	`json:"system" gorm:"default:false"` // дефолтный атрибут или нет (дефолтный нельзя удалить)
	EavAttrGroups 	[]EavAttrGroup 	`json:"eav_attr_groups" gorm:"many2many:eav_attr_eav_attr_group;"`
}

type EavAttrVarchar struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	EavAttrID 	uint 	`json:"-" gorm:"not null;"`
	Value 		string 	`json:"name" gorm:"not null;"` // "Тип сырья", "Цвет", "Размер", "Вес"

}

// таблица групп наборов атрибутов, на них ссылаются категории и товары.
type EavAttrGroup struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Name 		string 	`json:"name" gorm:"not null;"` // "Тип сырья", "Цвет", "Размер", "Вес"
	EavAttrs 	[]EavAttr 	`json:"eav_attrs" gorm:"many2many:eav_attr_eav_attr_group;"`
}

func (EavAttr) TableName() string {
	return "eav_attr"
}

func (EavAttrVarchar) TableName() string {
	return "eav_attr_varchar"
}
func (EavAttrGroup) TableName() string {
	return "eav_attr_group"
}


func (e *EavAttr) CreateHashID() error {
	return nil
}

func init() {


	/*db := base.GetDB()

	db.DropTableIfExists("eav_attr_eav_attr_groups")
	db.DropTableIfExists("eav_attr_group")
	db.DropTableIfExists("eav_attr_varchar")
	db.DropTableIfExists("eav_attrs")*/


	//db.DropTableIfExists(&EavAttrGroup{},&EavAttrVarchar{},&EavAttr{})


	//db.AutoMigrate(&EavAttrVarchar{})
	//db.AutoMigrate(&EavAttr{})
	//db.AutoMigrate(&EavAttrGroup{})

	//db.Table("eav_attrs").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	//db.Table("eav_attr_varchar").AddForeignKey("eav_attr_id", "eav_attrs(id)", "CASCADE", "CASCADE")

	//db.Table("eav_attr_eav_attr_group").AddForeignKey("eav_attr_id", "eav_attrs(id)", "CASCADE", "CASCADE")

}