package cluster

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/markgregr/bestHack_support_REST_server/internal/rest/forms"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
	log "github.com/sirupsen/logrus"
	"io"
)

type UpdateClusterRequest struct {
	Name string `json:"name"`
}

type UpdateClusterForm struct {
	Name string
}

func NewUpdateClusterForm() *UpdateClusterForm {
	return &UpdateClusterForm{}
}

func (f *UpdateClusterForm) ParseAndValidate(c *gin.Context) (forms.Former, response.Error) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()

	if err != nil {
		log.WithError(err).Error("unable to read body")
		return nil, response.NewInternalError()
	}

	var request *UpdateClusterRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetName(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *UpdateClusterForm) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"name": f.Name,
	}
}

func (f *UpdateClusterForm) validateAndSetName(request *UpdateClusterRequest, errors map[string]response.ErrorMessage) {
	if request.Name == "" {
		errors["name"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.Name = request.Name
}
