package behaviors

import (
	"github.com/derekyu332/goii/frame/base"
	"github.com/gin-gonic/gin"
)

type QueryFilter struct {
	Validators map[string][]base.IValidator
	//RulesMap map[string][]string
}

func (this *QueryFilter) BeforeAction(c *gin.Context) error {
	queries := c.Request.URL.Query()
	forms := c.Request.PostForm
	values := make(map[string]string)

	for key, value := range queries {
		if len(value) > 0 {
			values[key] = value[0]
		}
	}

	for key, value := range forms {
		if len(value) > 0 {
			values[key] = value[0]
		}
	}

	validators, ok := this.Validators[c.Request.URL.Path]

	if ok {
		for _, validator := range validators {
			if err := validator.Validate(values); err != nil {
				return base.InvalidParaHttpError(c, err.Error())
			}
		}
	}

	return nil
}

func (this *QueryFilter) AfterAction(c *gin.Context) error {
	return nil
}
