package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type ApiKey struct {
	ID uint `json:"id"`
	Token string `json:"token"` // varchar(32)
	AccountID uint `json:"-"`
	Name string `json:"name"`
	Status bool `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ### CRUD FUNC ###

// Создает нового пользователя с новым ID
func (a *ApiKey) create () error {

	//guid := xid.New()
	//a.Token = guid.String()
	a.Token = uuid.Must(uuid.NewV4()).String()


	//fmt.Println(uuid.Must(uuid.NewV4()))
	//fmt.Println( strings.ToLower(ksuid.New().String()))


	return db.Create(a).Error
}

// осуществляет поиск по ID
func (a *ApiKey) Get () error {
	return db.First(a,a.ID).Error
}

// сохраняет все поля в модели, кроме id, token, account_id, deleted_at
func (a *ApiKey) Save () error {
	return db.Model(User{}).Omit("id", "token","account_id","deleted_at").Save(a).Find(a, "id = ?", a.ID).Error
}

// обновляет все схожие с интерфейсом поля, кроме id, token, deleted_at
func (a *ApiKey) Update (input interface{}) error {
	return db.Model(User{}).Where("id = ?", a.ID).Omit("id", "token","account_id","deleted_at").Update(input).Find(a, "id = ?", a.ID).Error
}

// удаляет пользователя по ID
func (a *ApiKey) Delete () error {
	return db.Model(User{}).Where("id = ?", a.ID).Delete(a).Error
}