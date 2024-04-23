package forms

import (
	"github.com/gin-gonic/gin"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
)

type Former interface {
	ParseAndValidate(c *gin.Context) (Former, response.Error)
	ConvertToMap() map[string]interface{}
}
