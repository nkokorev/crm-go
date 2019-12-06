package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/utils"
)

// Product card
type Product struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`

	Article string `json:"article"`
	Name string `json:"name"`

	Account Account `json:"-"`

	Offers []ProductOffer `json:"offers"`
}

func (p *Product) create() error {

	// провекра, что такого
	if !db.Unscoped().First(&Product{},"sku = ?", p.Article).RecordNotFound() {
		return utils.Error{Message: fmt.Sprintf("Продукт с article = [%v] уже существует",p.Article) }
	}

	// создаем продукт
	if err := db.Create(p).Error; err != nil {
		return err
	}

	// тут надо создать офферы
	p.Offers = make([]ProductOffer, 0)

	// Создаем свойства (охуительная задача)
	//if len(p.Properties) > 0 {
	//	for _,r := range p.Properties {
	//		if err := r.create(); err != nil {
	//			fmt.Println("Cant create property product")
	//		}
	//		//fmt.Printf("Properties[%v] code: %v, value: %v \r\n", i, r.Code, r.Value)
	//	}
	//}

	return nil
}

// осуществляет поиск по Token
func (p *Product) get () error {
	// тут будет .Preload()
	return db.First(p, p.ID).Error
}

func (p *Product) save() error {
	// тут будет сохранение связанных данных
	return db.Model(p).Omit("id","account_id","deleted_at").Save(p).Find(p, p.ID).Error
}

// обновляет все схожие с интерфейсом поля, кроме id, account_id, deleted_at
func (p *Product) update (input interface{}) error {
	return db.Model(p).Where("id = ?", p.ID).Omit("id", "account_id", "deleted_at").Update(input).Find(p, "id = ?", p.ID).Error
}

func (p *Product) delete() error {
	return db.Model(p).Where("id = ?", p.ID).Delete(p).Error
}

// ### Управление моделью Модель работы работы с атрибутами

//// Добавление атрибута продукту (без значения)
//func (p *Product) CreateProperty(attr *Property) error {
//
//	return nil
//}
//
//// Получает данные по связанному объекту Property
//func (p Product) GetProperty(attr *Property) error {
//
//	return nil
//}
//
//// Обновленые каких-то данных связанности ... (каких?)
//func (p Product) UpdateProperty(attr *Property) error {
//
//	return nil
//}
//
//// убирает свойство у продукта
//func (p Product) DeleteProperty(attr *Property) error {
//
//	return nil
//}
//
//// Узнает, имеет ли продукт указанное свойство
//func (p Product) ExistProperty(code string) bool {
//
//	return false
//}
//
//// ### Работа с данными атрибутов
//
//// Задает новое значение свойства
//func (p *Product) SetPropertyValue(attr *Property) error {
//	return nil
//}
//
//// Возвращает пустую строку / массив, если такого объекта нет
//func (p *Product) GetPropertyValue(attr *Property) error {
//
//	return nil
//}