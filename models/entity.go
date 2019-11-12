package models

import (
	"errors"
	"github.com/nkokorev/crm-go/database/base"
)

// Блок по работе с моделями из аккаунта

// Entity: Product, Shop, Store, Role (?)

// в основе каждой Entity модели должны быть ниже приведенные закрытые (CRUD) методы
// каждая Entity модель явно привязана к Аккаунту. может этом как-то изменим и сделаем доп. методы без привязки или функцию чекалку привязки
// для внесения публичных изменений доступны публичные специфичные для каждой конкретной модели методы
type Entity interface {

	getID() uint
	getAccountID() uint
	//getHashID() string
	setAccountID(uint)

	// CRUD model
	create() error
	get (string) error // read ;)
	delete() error
	update() error

}

// создает модельку CRUD-методом
func (a Account) CreateEntity(v Entity) error {
	v.setAccountID(a.ID)
	return v.create()
}

// удаляет модельку CRUD-методом
func (a Account) DeleteEntity(v Entity) error {

	// проверяем принадлежность к аккаунту
	if ! a.isBelong(v) { return errors.New("This entity not belong current account") }

	// удаляем модельку crud-методом
	return v.delete()
}

// ищет по hash_id принадлежающую аккаунту модель связанную с аккаунтом или ошибку, если модель не найдена или не принадлежит этому аккаунту
func (a Account) GetEntity (v Entity, hash string) error {
	// получаем саму модель встроенным в модель методом
	if err := v.get(hash); err != nil {return err}

	// проверяем принадлежность к аккаунту
	if ! a.isBelong(v) { return errors.New("This entity not belong current account") }

	return nil
}

// проверяет наличие модели и принадлежность к аккаунту
func (a Account) HasEntity (v Entity, hash string) bool {
	if v.get(hash) != nil { return true }
	return false
}

// проверяет принадлежность модели к аккаунту
func (a Account) isBelong(v Entity) bool {
	return v.getAccountID() == a.ID
}

// эксперементальная функция, filter => Where(filter...)
func (a Account) GetEntities (vals interface{}, filter... interface{}) error {

	switch len(filter) {
	case 0:
		return base.GetDB().Where("account_id = ?", a.getID()).Find(vals).Error;
	case 1:
		return base.GetDB().Where("account_id = ?", a.getID()).Where(filter[0]).Find(vals).Error;
	case 2:
		return base.GetDB().Where("account_id = ?", a.getID()).Where(filter[0], filter[1]).Find(vals).Error;
	default:
		return base.GetDB().Where("account_id = ?", a.getID()).Where(filter).Find(vals).Error;
	}

}

// эксперементальная функция по множественному созданию сущностей
func (a Account) CreateEntities (vals []Entity) error {
	for i,_ := range vals {
		if err := a.CreateEntity(vals[i]); err != nil {
			return err
		}
	}
	return nil
}