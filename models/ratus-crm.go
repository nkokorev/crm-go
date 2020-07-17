package models

// Системные функции
type RatusCRM struct {

}

// Разрешение пользователю другого аккаунта входить через https://app.ratuscrm.com/login
func (RatusCRM) AllowedUserLoginCRM(userId uint) error {

	user, err := getUserById(userId)
	if err != nil {
		return err
	}
	user.EnabledAuthFromApp = true
	if err = user.save(); err != nil {
		return err
	}

	return nil
}