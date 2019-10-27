package models

import (
	"errors"
	"github.com/nkokorev/crm-go/database/base"
	e "github.com/nkokorev/crm-go/errors"
	u "github.com/nkokorev/crm-go/utils"
	"reflect"
	"time"
)

type ApiKey struct {
	ID			uint `json:"-" gorm:"primary_key;unique_index;"`
	Token 		string `json:"hash_id" gorm:"unique_index;varchar(32)"` // сам ключ доступа длиной в 32 символа
	AccountID 	uint `json:"-" gorm:"index;"` // Owner token, foreignKey !
	Name 		string `json:"name" gorm:"size:255"` // Назначение: 'Токен для сайта', 'Для тестовых подключений'
	Status 		bool `json:"status" gorm:"default:true"` // статус ключа (активирован ли)

	// todo: надо доработать, чтобы модель подгружала сразу нужные данные с ролями и владельцами
	Role		Role `gorm:"foreignkey:RoleID;default:NULL"`
	RoleID   	uint `json:"role_id"` // роль определяет уровнь доступа токена
	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
}

// создает apiKey с токеном доступа
func (key *ApiKey) create() error {

	// проверка повторного создания ключа
	if reflect.TypeOf(key.ID).String() == "uint" {
		if key.ID > 0 && !base.GetDB().First(&ApiKey{}, key.ID).RecordNotFound() {
			return errors.New("Can't create new api-key: already crated!")
		}
	}

	// создаем уникальный токен
	if err := key.createToken();err != nil {
		return err
	}
	// создаем api-ключ в БД
	if err := base.GetDB().Create(key).Error; err != nil {
		return err
	}
	// фиксим баг с timestamps
	if err := base.GetDB().Save(key).Error; err != nil {
		return err
	}

	return nil
}

// удаляет ключ
func (key *ApiKey) delete() error {

	if reflect.TypeOf(key.ID).String() != "uint" {
		return e.RoleDeletionError
	}

	if err := base.GetDB().Unscoped().Delete(key).Error; err != nil {
		return err
	}

	// возможно тут нужно обнулять данные модели..
	return nil
}

// устанавливает новую роль api ключу.
func (key *ApiKey) SetRole(role *Role) error {

	if err := base.GetDB().Model(key).Update("RoleID", role.ID).Error; err != nil {
		return err
	}
	return nil
}

// ниже вспомогательные функции простановки системных ролей api-ключам
func (key *ApiKey) SetRoleFullAccess() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.FindRoleByTag("full-access");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := key.SetRole(&role);err != nil {
		return err
	}
	return nil
}
func (key *ApiKey) SetRoleSiteAccess() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.FindRoleByTag("site-access");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := key.SetRole(&role);err != nil {
		return err
	}
	return nil
}
func (key *ApiKey) SetRoleReadAccess() error {
	// 1. Ищем необходимую роль
	role := Role{}
	if err := role.FindRoleByTag("read-access");err != nil {
		return err
	}

	// 2. Ставим пользователю найденную роль
	if err := key.SetRole(&role);err != nil {
		return err
	}
	return nil
}

func (key *ApiKey) createToken() error {

	token := u.RandStringBytes(32)

	// проверям, что такого токена нет
	if !base.GetDB().First(&ApiKey{}, "token = ?",token).RecordNotFound() {
		return errors.New("Cant create token!")
	}

	key.Token = token

	return nil
}