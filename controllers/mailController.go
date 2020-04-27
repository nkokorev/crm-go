package controllers

import (
    "bytes"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/asn1"
    "encoding/json"
    "encoding/pem"
    "fmt"
    "github.com/nkokorev/crm-go/models"
    u "github.com/nkokorev/crm-go/utils"
    "html/template"
    "log"
    "net/http"
    "net/mail"
    "os"
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

func Keymaker(path string) {
    reader := rand.Reader
    bitSize := 2048

    key, err := rsa.GenerateKey(reader, bitSize)
    checkError(err)

    publicKey := key.PublicKey

    //saveGobKey("private.key", key)
    savePEMKey(path+".priv", key)

    //saveGobKey("public.key", publicKey)
    savePublicPEMKey(path+".pub", publicKey)
}

func savePEMKey(fileName string, key *rsa.PrivateKey) {
    outFile, err := os.Create(fileName)
    checkError(err)
    defer outFile.Close()

    var privateKey = &pem.Block{
        Type:  "PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(key),
    }

    err = pem.Encode(outFile, privateKey)
    checkError(err)
}

func savePublicPEMKey(fileName string, pubkey rsa.PublicKey) {
    asn1Bytes, err := asn1.Marshal(pubkey)
    checkError(err)

    var pemkey = &pem.Block{
        Type:  "PUBLIC KEY",
        Bytes: asn1Bytes,
    }

    pemfile, err := os.Create(fileName)
    checkError(err)
    defer pemfile.Close()

    err = pem.Encode(pemfile, pemkey)
    checkError(err)
}

func checkError(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }
}

func GenerateRsaKey() {

    pKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        //u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"pkey":err.Error()}}))
        log.Fatal(err)
        return
    }
    //fmt.Println(pKey )
    fmt.Println(pKey.PublicKey.Size() )
    //key :=
    //fmt.Println(x509.MarshalPKCS1PrivateKey(key))

}
