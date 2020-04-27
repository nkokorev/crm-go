package controllers

import (
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/nkokorev/crm-go/models"
    u "github.com/nkokorev/crm-go/utils"
    "html/template"
    "log"
    "net/http"
    "net/mail"
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

    input := struct {
        To string `json:"to"`
        FromAddress string `json:"fromAddress"`
        FromName string `json:"fromName"`
        Subject string `json:"subject"`
    }{}

    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке json-запроса"}))
    }
    
    // делаем тестовую отправку письма

    go func() {
        //tpl, err := template.ParseFiles("files/example.html")
        tpl, err := template.ParseFiles("files/example.html")
        if err != nil {
            log.Fatal(err)
            return
        }

        buf := bytes.Buffer{}
        if err := tpl.Execute(&buf, struct {
            Title string
            Name string
        }{"Test msg", "Nikita"}); err != nil {
            log.Fatal(err)
            return
        }

        message := models.Message{
            To: mail.Address{Name: "", Address: input.To},
            From: mail.Address{Name: input.FromName, Address: input.FromAddress},
            Subject: input.Subject,
            Body: buf.String(),
        }
        if err := message.Send(); err != nil {
            fmt.Println("Cant sent message")
            log.Fatal(err)
        } else {
            fmt.Println("Msg sent")
        }
    }()


    resp := u.Message(true, "Message sent successful!")
    u.Respond(w, resp)
}
