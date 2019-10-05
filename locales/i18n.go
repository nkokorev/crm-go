package locales

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle = i18n.NewBundle(language.Russian)
var localizer = i18n.NewLocalizer(bundle)

func init() {
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	filesToml := []string {"user", "active"} // file's: locales.user.ru.toml
	for _, value := range filesToml {
		bundle.MustLoadMessageFile("locales/toml/" + value + ".ru.toml")
		bundle.MustLoadMessageFile("locales/toml/" + value + ".en.toml")
	}

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
func Trans(Id string, b_optional ...int) string {

	PluralCount := 1
	if len(b_optional) > 0 {
		PluralCount = b_optional[0]
	}

	resp, err := GetLocalizer().Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    Id,
		},
		PluralCount: PluralCount,
	})
	if err != nil {
		return ""
	}

	return resp
}

