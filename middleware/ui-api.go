package middleware

import (
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)


// Public UI-API https://ui.api.ratuscrm.com
func UiApiPublicAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Подгрузим файл настроек, он тут будет кстати
		crmSettings, err := models.CrmSetting{}.Get()
		if err != nil {
			u.Respond(w, u.MessageError(nil, "Сервер не может обработать запрос")) // что это?)
			return
		}

		// если Public UI-API погашен
		if !crmSettings.UiApiPublicEnabled {
			u.Respond(w, u.MessageError(nil, crmSettings.UiApiDisabledMessage)) // что это?)
			return
		}


		/*ctx1 := context.WithValue(r.Context(), "user_id", tk.UserId)
		r = r.WithContext(ctx1)

		ctx2 := context.WithValue(r.Context(), "account_id", tk.AccountId)
		r = r.WithContext(ctx2)*/

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	});
}
