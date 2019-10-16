package locales

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var bundle = i18n.NewBundle(language.Russian)
var localizer = i18n.NewLocalizer(bundle)

func init() {
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
}

func LoadFromPathFolder(dir string) {

	var files []string

	root, errf := filepath.Abs(filepath.Dir(dir))
	if errf != nil {
		log.Fatal(errf)
		return
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".toml" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		bundle.MustLoadMessageFile(file)
		//fmt.Println("Load localized file: " + file)
	}


	//filesToml := []string {"user", "active"} // file's: locales.user.ru.toml
	//for _, value := range filesToml {
	//	path := "locales/toml/" + value + ".ru.toml"
	//	dir, err := filepath.Abs(path)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	//fmt.Println(dir)
	//	//bundle.MustLoadMessageFile("locales/toml/" + value + ".ru.toml")
	//	//bundle.MustLoadMessageFile("locales/toml/" + value + ".en.toml")
	//}

}

func GetBundle() *i18n.Bundle {
	return bundle
}

func GetLocalizer() *i18n.Localizer {
	return localizer
}

func SetAccept (accept string) {
	localizer = i18n.NewLocalizer(bundle, accept)
}

// Id - from /locales.const.go
func Trans(Id string, b_optional ...map[string]interface{} ) string {

	var Data map[string]interface{}
	PluralCount := 1

	if len(b_optional) > 0 {
		Data = b_optional[0]
		if _, ok := b_optional[0]["PluralCount"]; ok {
			var errp error
			PluralCount, errp = strconv.Atoi( fmt.Sprintf("%d", b_optional[0]["PluralCount"]) )
			if errp != nil {
				log.Println("Trans: wrong PluralCount variable: ", PluralCount)
				PluralCount = 1
			}
		}
	}

	resp, err := GetLocalizer().Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    Id,
		},
		TemplateData: Data,
		PluralCount: PluralCount,
	})
	if err != nil {
		return ""
	}

	return resp
}

