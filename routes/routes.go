package routes

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/nkokorev/crm-go/controllers"
	//"github.com/nkokorev/gui-server/controllers/old"
	"github.com/nkokorev/crm-go/middleware"
)

func Handlers() *mux.Router {

	// обрабатываем все запросы со слешем и без
	//r := mux.NewRouter().StrictSlash(false)
	r_base := mux.NewRouter()
	r_base = r_base.PathPrefix("/api").Subrouter()

	// добавляем в ответ Access-Control-Allow-Origin и прочие заголовки для выделенных серверов (dev,local & production servers)
	r_base.Use(middleware.CorsAccessControl)

	r_user := r_base.PathPrefix("").Subrouter()
	r_user.Use(middleware.JwtUserAuthentication)

	r_acc := r_base.PathPrefix("").Subrouter()
	r_acc.Use(middleware.JwtAccountAuthentication)

	r_full := r_base.PathPrefix("").Subrouter()
	r_full.Use(middleware.JwtFullAuthentication)

	// Группы маршрутов по моделям
	UserRoutes(r_base.PathPrefix("/user").Subrouter(),
		r_user.PathPrefix("/user").Subrouter(),
		r_acc.PathPrefix("/user").Subrouter(),
		r_full.PathPrefix("/user").Subrouter(),
	)

	// Lll
	AppRoutes(
		r_base.PathPrefix("/app").Subrouter(),
		r_user.PathPrefix("/app").Subrouter(),
		r_acc.PathPrefix("/app").Subrouter(),
		r_full.PathPrefix("/app").Subrouter(),
	)

	// SS
	AccountRoutes(r_base.PathPrefix("/accounts").Subrouter(),
		r_user.PathPrefix("/accounts").Subrouter(),
		r_acc.PathPrefix("/accounts").Subrouter(),
		r_full.PathPrefix("/accounts").Subrouter(),
	)

	r_base.NotFoundHandler = middleware.NotFoundHandler()

	return r_base
}

/**
* [/app] - группа роутов отвечающих за авторизацию и аутентификацию пользователя и аккаунта
 */
var AppRoutes = func (r_base, r_user, r_acc, r_full *mux.Router) {

	// загружаем базовые настройки системы
	r_base.HandleFunc("/settings", controllers.CrmGetSettings).Methods(http.MethodGet, http.MethodOptions)


	r_user.HandleFunc("/auth/user", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodOptions)
	r_acc.HandleFunc("/auth/account", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodOptions)
	r_full.HandleFunc("/auth", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodOptions)

}

/**
* [/user] - группа роутов отвечающих за создание пользователей
 */
var UserRoutes = func (r_base, r_user, r_acc, r_full *mux.Router) {

	// hidden: code for auth: account_name: <>
	r_base.HandleFunc("/auth", controllers.UserAuthorization).Methods(http.MethodPost, http.MethodOptions)

	r_base.HandleFunc("", controllers.UserCreate).Methods(http.MethodPost, http.MethodOptions)
	r_user.HandleFunc("", controllers.UserGetProfile).Methods(http.MethodGet, http.MethodOptions)

	r_base.HandleFunc("/email-verification", controllers.UserEmailVerificationConfirm).Methods(http.MethodPost, http.MethodOptions)

	//r_user.HandleFunc("/email-verification", crm.UserSendEmailVerification).Methods(http.MethodGet, http.MethodOptions)
	r_user.HandleFunc("/email-verification/invite", controllers.UserSendEmailInviteVerification).Methods(http.MethodGet, http.MethodOptions)

	r_base.HandleFunc("/recovery/username", controllers.UserRecoveryUsername).Methods(http.MethodPost, http.MethodOptions)

	// ### Password -Routes- ###

	// Устанавливает новый пароль {"password":"...", "password_old":"..."}. Если password_reset = true, то старый пароль не требуется
	r_user.HandleFunc("/password", controllers.UserSetPassword).Methods(http.MethodPost, http.MethodOptions)

	// отправляет email с инструцией по сбросу пароля
	r_base.HandleFunc("/password/reset/send-email", controllers.UserRecoveryPasswordSendMail).Methods(http.MethodPost, http.MethodOptions)

	// одноразовая авторизация по {"token":"<...>"}: сбрасывает пароль и совершает авторизацию.
	r_base.HandleFunc("/password/reset/confirm", controllers.UserPasswordResetConfirm).Methods(http.MethodPost, http.MethodOptions)


	// отмена сброса пароля. Важная функция для антивзлома. Можно использовать тот же токен, но лучше создавать отдельный...
	//r_base.HandleFunc("/password/reset/report", crm.UserRecoveryPassword).Methods(http.MethodDelete, http.MethodOptions)

	// почему бы функции ниже не перенести в [AccountRoutes]?
	r_user.HandleFunc("/accounts", controllers.UserGetAccounts).Methods(http.MethodGet, http.MethodOptions)
	r_user.HandleFunc("/accounts/{account_id:[0-9]+}/auth", controllers.UserLoginInAccount).Methods(http.MethodGet, http.MethodOptions)

}

/**
* [/accounts] - группа роутов отвечающих за авторизацию и аутентификацию пользователя и аккаунта
 */
var AccountRoutes = func (r_base, r_user, r_acc, r_full *mux.Router) {

	r_user.HandleFunc("", controllers.AccountCreate).Methods(http.MethodPost, http.MethodOptions)
	r_acc.HandleFunc("", controllers.AccountGetProfile).Methods(http.MethodGet, http.MethodOptions)

}