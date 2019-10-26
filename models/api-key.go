package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"time"
)

type ApiKey struct {
	ID			uint `gorm:"primary_key;unique_index;" json:"-"`
	Token 		string `json:"hash_id" gorm:"unique_index;varchar(32)"`
	AccountID 	uint `json:"account_id" gorm:"index;"` // Owner token, foreignKey !
	UserID 		string `json:"creator_id"` // кто создал, можно потом привязать к hash_id
	Label 		string `json:"name" gorm:"size:255"` // 'Токен для сайта'
	Status 		bool `json:"status"`
	RoleID   	uint `json:"role_id"`
	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
}

// устанавливает новую роль api ключу.
func (key *ApiKey) SetNewRole(role *Role) error {

	if err := base.GetDB().Model(key).Update("RoleID", role.ID).Error; err != nil {
		return err
	}
	return nil
}

// ниже вспомогательные функции простановки системных ролей api-ключам
func (key *ApiKey) SetRoleFullAssets() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.GetRoleByTag("full-assets");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := key.SetNewRole(&role);err != nil {
		return err
	}
	return nil
}
func (key *ApiKey) SetRoleSiteAccess() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.GetRoleByTag("site-access");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := key.SetNewRole(&role);err != nil {
		return err
	}
	return nil
}
func (key *ApiKey) SetRoleReadAccess() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.GetRoleByTag("read-access");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := key.SetNewRole(&role);err != nil {
		return err
	}
	return nil
}