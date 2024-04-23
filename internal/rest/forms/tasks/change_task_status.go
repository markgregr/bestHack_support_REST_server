package tasks

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/markgregr/bestHack_support_REST_server/internal/rest/forms"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
	log "github.com/sirupsen/logrus"
	"io"
)

type ChangeTaskStatusRequest struct {
	TaskID int64 `json:"task_id"`
}

type ChangeTaskStatusForm struct {
	TaskID int64
}

func NewChangeTaskStatusForm() *ChangeTaskStatusForm {
	return &ChangeTaskStatusForm{}
}

func (f *ChangeTaskStatusForm) ParseAndValidate(c *gin.Context) (forms.Former, response.Error) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()

	if err != nil {
		log.WithError(err).Error("unable to read body")
		return nil, response.NewInternalError()
	}

	var request *ChangeTaskStatusRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}
	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetTitle(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *ChangeTaskStatusForm) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"task_id": f.TaskID,
	}
}

func (f *ChangeTaskStatusForm) validateAndSetTitle(request *ChangeTaskStatusRequest, errors map[string]response.ErrorMessage) {
	if request.TaskID == 0 {
		errors["task_id"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.TaskID = request.TaskID
}
