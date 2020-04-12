package controllers

import (
    "fmt"
    "github.com/nkokorev/crm-go/models"
    u "github.com/nkokorev/crm-go/utils"
    "net/http"
)

func SendEmailMessage(w http.ResponseWriter, r *http.Request) {

    if r.Context().Value("account") == nil {
        u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"account":"not load"}}))
        return
    }
    account := r.Context().Value("account").(*models.Account)
    if &account == nil {
        fmt.Println("Account is not found!")
        u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"account":"not load"}}))
        return
    }

    /*fmt.Println("issuer acc: ", r.Context().Value("issuerAccountId"))
    fmt.Println("account id: ", r.Context().Value("accountId"))
    fmt.Printf("Account is: %v \n", account.Name)*/
    
    resp := u.Message(true, "Message sent successful!")
    u.Respond(w, resp)
}
