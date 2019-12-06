package models

type EavAttrType struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Table string `json:"table_name" gorm:"table_name"`
	Description string `json:"description"`
}

func (EavAttrType) TableName() string {
	return "eav_attr_type"
}
