package models

type Product struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`
	SKU string `json:"sku"`
	Name string `json:"name"`
	
	Account Account `json:"-"`
}

func (p *Product) create() error {
	return db.Create(p).Error
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

