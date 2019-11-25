package models

import "time"

type ApiKey struct {
	ID uint `json:"id"`
	Token string `json:"token"` // varchar(32)
	AccountID uint `json:"-"`
	Name string `json:"name"`
	Status bool `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *ApiKey) create () error {
	return nil
}
func (a *ApiKey) get () error {
	return nil
}
func (a *ApiKey) save () error {
	return nil
}
func (a *ApiKey) delete () error {
	return nil
}