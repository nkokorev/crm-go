package models

import (
	"errors"
)

// Блок по работе с моделями из аккаунта

// Entity: Product, Shop, Store, Role (?)

// в основе каждой Entity модели должны быть ниже приведенные закрытые (CRUD) методы
// каждая Entity модель явно привязана к Аккаунту. может этом как-то изменим и сделаем доп. методы без привязки или функцию чекалку привязки
// для внесения публичных изменений доступны публичные специфичные для каждой конкретной модели методы
type Entity interface {

	getID() uint
	getAccountID() uint
	getHashID() string
	setAccountID(uint)

	// CRUD model
	create() error
	getByHashID (string) error // read ;)
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
func (a Account) GetEntity (hash string, v Entity) error {
	// получаем саму модель встроенным в модель методом
	if err := v.getByHashID(hash); err != nil {return err}

	// проверяем принадлежность к аккаунту
	if ! a.isBelong(v) { return errors.New("This entity not belong current account") }

	return nil
}

// проверяет наличие модели и принадлежность к аккаунту
func (a Account) HasEntity (hash string, v Entity) bool {
	if v.getByHashID(hash) != nil { return true }
	return false
}

// проверяет принадлежность модели к аккаунту
func (a Account) isBelong(v Entity) bool {
	return v.getAccountID() == a.ID
}
