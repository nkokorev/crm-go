package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/middleware"
	"os"
)

func Handlers() *mux.Router {

	var crmHost string

	// root route - handle all request
	r := mux.NewRouter().StrictSlash(true)

	// Environment variable: local, public etc..
	AppEnv := os.Getenv("APP_ENV")

	// Set AppEnv variable
	switch AppEnv {
	case "local":
		crmHost = "crm.local"
		//crmHost = "127.0.0.1:8090"
		//crmHost = "localhost:8090"
	case "public":
		crmHost = "ratuscrm.com"
	default:
		crmHost = "ratuscrm.com"
	}

	// Set CORS politic for local development
	if AppEnv == "local" {
		r.Use(middleware.CorsAccessControl)
	}

	// Mount all root point of routes
	rApp := r.Host("app." + crmHost).PathPrefix("/ui-api").Subrouter() // APP [app.ratuscrm.com/ui-api]
	rApi := r.Host("api." + crmHost).Subrouter() // API [api.ratuscrm.com]
	rUiApi := r.Host("ui.api." + crmHost).PathPrefix("/accounts/{accountHashId:[a-z0-9]+}").Subrouter() // UI/API [ui.api.ratuscrm.com]
	rCDN := r.Host("cdn." + crmHost).Subrouter() // API [cdn.ratuscrm.com]
	rTracking := r.Host("tracking." + crmHost).Subrouter() // API [tracking.ratuscrm.com]
	rMTA1 := r.Host("mta1." + crmHost).Subrouter() // API [mta1.ratuscrm.com]

	// rShare := r.Host("share." + crmHost).Subrouter() // API [share.ratuscrm.com]


	/******************************************************************************************************************

		###	Authentication rules ###

		1. В App базовая авторизация производится в RatusCRM аккаунте.
		2. На этапе выдачи account-token issuerAccount становится равным аккаунту, в котором авторизован пользователь.
		3. AuthToken: RatusCRM => IssuerAccount

		### Context(r) ###

		1. issuerAccount - auth Account | (*models.Account), example: issuerAccount.DecryptToken(token string) (error, token)
		2. account - target (load/work) Account | (*models.Account), example: account.LoginUser(username, password string)
		3. user - auth user in work account | (*models.User). Example: user.CreateAccount(Account{...})
		4. issuer - channel of request: 'app', 'api', 'ui-api'. Need for some logic controllers (^_^)

		* App(r): decrypting with RatusCRM AES/JWT key. Main account adding to context(r) issuerAccount
		* UI(p): decrypting with Account AES/JWT key. Account add to context(r) issuerAccount from .../{account_hash_id}/...
		* API(r): getting Bearer token from Headers(r) and compare in api-key table

		### Middleware(r) ###

		1. middleware.CheckApiStatus - проверяет статус API для всех аккаунтов
		2. middleware.CheckAppUiApiStatus - проверяет статус App UI/API в настройках GUI RatusCRM
		3. middleware.CheckUiApiStatus - проверяет статус Public UI/API для всех аккаунтов

		4. middleware.ContextMuxVarAccount - Вставляет в контекст issuerAccountId из hashAccountId (раскрытие issuer account) && issuerAccount
		5. middleware.ContextMainAccount - устанавливает в контекст issuerAccountId = 1 && issuerAccount

	  	6. middleware.BearerAuthentication - читает с проверкой JWT, проверяет статус API в аккаунте. Дополняет контекст accountId && account
		7. middleware.JwtUserAuthentication - проверяет JWT и устанавливает в контекст userId & user
		8. middleware.JwtFullAuthentication - проверяет JWT и устанавливает в контекст userId & user, accountId && account

	******************************************************************************************************************/
	rApi.Use	(middleware.CorsAPIAccessControl, 		middleware.CheckApiStatus, 		middleware.BearerAuthentication)
	rApp.Use	(middleware.CheckAppUiApiStatus,	middleware.AddContextMainAccount)
	// rUiApi.Use	(middleware.CorsAccessControl, 		middleware.CheckUiApiStatus, 	middleware.ContextMuxVarAccountHashId)
	rUiApi.Use	(middleware.CorsAccessControl, 		middleware.CheckUiApiStatus, 	middleware.ContextMuxVarAccountHashId)

	// RouteHandlers
	AppRoutes(rApp)
	ApiRoutes(rApi)
	UiApiRoutes(rUiApi)
	CDNRoutes(rCDN)
	TrackingRoutes(rTracking)
	MTA_1_Routes(rMTA1)


	// ### 404 (^_^) ###
	r.NotFoundHandler = middleware.NotFoundHandler()

	return r
}
