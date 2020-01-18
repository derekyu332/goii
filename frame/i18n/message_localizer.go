package i18n

import (
	"github.com/BurntSushi/toml"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/gin-gonic/gin"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	gDefaultBundle *goi18n.Bundle
)

func InitBundle(messageFiles []string) error {
	gDefaultBundle = goi18n.NewBundle(language.English)
	gDefaultBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	for _, path := range messageFiles {
		gDefaultBundle.MustLoadMessageFile(path)
	}

	logger.Warning("i18n Bundle Init Success")
	logger.Warning("%v", gDefaultBundle.LanguageTags())

	return nil
}

func T(c *gin.Context, messageId string, templateData interface{}) string {
	if localizer := NewLocalizer(c, nil); localizer != nil {
		return L(localizer, &goi18n.LocalizeConfig{
			MessageID:    messageId,
			TemplateData: templateData,
		})
	} else {
		return ""
	}
}

func M(c *gin.Context, messageId string) string {
	if localizer := NewLocalizer(c, nil); localizer != nil {
		return L(localizer, &goi18n.LocalizeConfig{
			MessageID: messageId,
		})
	} else {
		return ""
	}
}

func L(localizer *goi18n.Localizer, lc *goi18n.LocalizeConfig) string {
	if localizer == nil || lc == nil {
		return ""
	}

	if s, err := localizer.Localize(lc); err != nil {
		return ""
	} else {
		return s
	}
}

func NewLocalizer(c *gin.Context, bundle *goi18n.Bundle) *goi18n.Localizer {
	if bundle == nil {
		bundle = gDefaultBundle

		if bundle == nil {
			return nil
		}
	}

	if c == nil {
		return goi18n.NewLocalizer(bundle, "en-US")
	} else {
		lang := c.Request.FormValue("lang")
		accept := c.Request.Header.Get("Accept-Language")
		localizer := goi18n.NewLocalizer(bundle, lang, accept)

		return localizer
	}
}
