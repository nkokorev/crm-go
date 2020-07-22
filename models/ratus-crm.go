package models

// Системные функции
type RatusCRM struct {

}

// Разрешение пользователю другого аккаунта входить через https://app.ratuscrm.com/login
func (RatusCRM) AllowedUserLoginCRM(userID uint) error {

	user, err := getUserByID(userID)
	if err != nil {
		return err
	}
	user.EnabledAuthFromApp = true
	if err = user.save(); err != nil {
		return err
	}

	return nil
}