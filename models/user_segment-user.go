package models

import (
	"log"
)

// HELPER FOR M<>M IN PgSQL
type UsersSegmentUser struct {
	UserId  		uint `json:"user_id" gorm:"type:int;index;not null;"`
	UsersSegmentId 	uint `json:"users_segment_id" gorm:"type:int;index;not null;"`
}

func (UsersSegmentUser) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&UsersSegmentUser{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE users_segment_users \n    ADD CONSTRAINT users_segment_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT users_segment_users_users_segment_id_fkey FOREIGN KEY (users_segment_id) REFERENCES users_segments(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_users_segment_users_users,\n    DROP CONSTRAINT IF EXISTS fk_users_segment_users_user_segments;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

}

func (UsersSegmentUser) TableName() string {
	return "users_segment_users"
}
