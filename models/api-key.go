package models

import (
	"errors"
	"github.com/nkokorev/crm-go/database/base"
	e "github.com/nkokorev/crm-go/errors"
	u "github.com/nkokorev/crm-go/utils"
	"reflect"
	"time"
)

// Support Account Entity
type ApiKey struct {
	ID			uint `json:"-" gorm:"primary_key;unique_index;"`
	Token 		string `json:"token" gorm:"unique_index;varchar(32)"` // сам ключ доступа длиной в 32 символа
	//Account		Role `json:"-" gorm:"foreignkey:AccountID;default:NULL"`
	AccountID 	uint `json:"-" gorm:"index;"` // Owner token, foreignKey !
	Name 		string `json:"name" gorm:"size:255"` // Назначение: 'Токен для сайта', 'Для тестовых подключений'
	Status 		bool `json:"status" gorm:"default:true"` // статус ключа (активирован ли)

	// todo: надо доработать, чтобы модель подгружала сразу нужные данные с ролями и владельцами
	Role		Role `gorm:"foreignkey:RoleID;default:NULL"`
	RoleID   	uint `json:"role_id" gorm:"default:NULL"` // роль определяет уровнь доступа токена
	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
}

// вспомогательная функция для получения ID
func (k ApiKey) getID () uint { return k.ID }

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

// обновляет данные ключа с защитой служебных полей
func (key *ApiKey) update() (err error) {

	// указываем какие поля обновлять не надо
	if err := base.GetDB().Model(&key).Omit("id", "hash_id", "account_id").Updates(key).Error; err != nil {
		return err
	}

	return nil
}

// устанавливает новую роль api ключу.
func (key *ApiKey) SetRole(role *Role) error {

	if err := base.GetDB().Model(key).Update("RoleID", role.ID).Error; err != nil {
		return err
	}
	return nil
}

// получить роль пользователя
func (key *ApiKey) GetRole(role *Role) error {

	// Найдем роль пользователя
	if err := base.GetDB().Model(key).Related(role).Error;err != nil {
		return err
	}
	return nil
}

// Проверяем права доступа для конкретного api-ключа.
func (key *ApiKey) CheckPermission(permission_code uint) bool {

	// 1. Получаем роль доступа
	role := Role{}
	if err := key.GetRole(&role); err != nil {
		return false
	}

	// 2. Если пользователь имеет роль владельца аккаунта (или аналогичный ключ) - сразу даем зеленый свет - это сильно ускоряет работу.
	if role.Tag == "full-access" {
		return true
	}

	// Если роль отличная от full-access, то сделаем прямой запрос в БД через функцию Роли
	if role.hasPermission(permission_code) {
		return true
	}

	// не нашли доказательств наличия права
	return false
}

// ниже вспомогательные функции простановки системных ролей api-ключам
func (key *ApiKey) SetRoleFullAccess() error {

	// 0. todo: Если ключ еще не создан, эта конструкция работает, но ее можно упростить

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

// вспомогательная функция для получения ID
func (k ApiKey) getAccountID () (id uint) { return k.AccountID }

// вспомогательная функция для получения ID
func (k *ApiKey) setAccountID (id uint) { k.AccountID = id }

// ищет продукт по hashID. Возвращает ошибку, если ключ не найден или еще что-то пошло не так
func (k *ApiKey) get(token string) error {

	if err := base.GetDB().First(k,"token = ?", token).Error;err != nil {
		return err
	}
	return nil
}

func (key *ApiKey) Get(token string) error {

	if err := base.GetDB().First(key,"token = ?", token).Error;err != nil {
		return err
	}
	return nil
}