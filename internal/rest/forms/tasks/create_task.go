package tasks

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/markgregr/bestHack_support_REST_server/internal/rest/forms"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
	log "github.com/sirupsen/logrus"
	"io"
)

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateTaskForm struct {
	Title       string
	Description string
}

func NewCreateTaskForm() *CreateTaskForm {
	return &CreateTaskForm{}
}

func (f *CreateTaskForm) ParseAndValidate(c *gin.Context) (forms.Former, response.Error) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()

	if err != nil {
		log.WithError(err).Error("unable to read body")
		return nil, response.NewInternalError()
	}

	var request *CreateTaskRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}
	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetTitle(request, errors)
	f.validateAndSetDescription(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *CreateTaskForm) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"title":       f.Title,
		"description": f.Description,
	}
}

func (f *CreateTaskForm) validateAndSetTitle(request *CreateTaskRequest, errors map[string]response.ErrorMessage) {
	if request.Title == "" {
		errors["title"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.Title = request.Title
}

func (f *CreateTaskForm) validateAndSetDescription(request *CreateTaskRequest, errors map[string]response.ErrorMessage) {
	if request.Description == "" {
		errors["description"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.Description = request.Description
}

//func (f *CreateTaskForm) validateAndSetClusterIndex(request *CreateTaskRequest, errors map[string]response.ErrorMessage) {
//	if request.ClusterIndex == 0 {
//		errors["cluster_index"] = response.ErrorMessage{
//			Code:    response.MissedValue,
//			Message: "missed value",
//		}
//		return
//	}
//
//	f.ClusterIndex = request.ClusterIndex
//}
