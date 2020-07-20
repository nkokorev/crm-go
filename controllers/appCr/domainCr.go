package appCr

import (
	"crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/asn1"
    "encoding/pem"
    "fmt"
    "github.com/nkokorev/crm-go/controllers/utilsCr"
    u "github.com/nkokorev/crm-go/utils"
    "log"
    "net/http"
    "os"
)

func DomainsGet(w http.ResponseWriter, r *http.Request) {

    account, err := utilsCr.GetWorkAccount(w,r)
    if err != nil || account == nil {
        u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
        return
    }

    domains, err := account.GetDomains()
    if err != nil || domains == nil {
        u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"domains":"Не удалось получить список доменов"}}))
        return
    }

    resp := u.Message(true, "GET account domains")
    resp["domains"] = domains
    u.Respond(w, resp)
}

func Keymaker(path string) {
    reader := rand.Reader
    // bitSize := 2048
    bitSize := 1024

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
        //Bytes: bytesKey,
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
