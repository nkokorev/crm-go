package models

import "time"

// public model
type DBModel struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}