package tasks

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/markgregr/bestHack_support_REST_server/internal/rest/forms"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
	log "github.com/sirupsen/logrus"
	"io"
)

type AddSolutionToTaskRequest struct {
	Solution string `json:"solution"`
}

type AddSolutionToTaskForm struct {
	Solution string
}

func NewAddSolutionToTaskForm() *AddSolutionToTaskForm {
	return &AddSolutionToTaskForm{}
}

func (f *AddSolutionToTaskForm) ParseAndValidate(c *gin.Context) (forms.Former, response.Error) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()

	if err != nil {
		log.WithError(err).Error("unable to read body")
		return nil, response.NewInternalError()
	}

	var request *AddSolutionToTaskRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}
	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetSolution(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *AddSolutionToTaskForm) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"solution": f.Solution,
	}
}

func (f *AddSolutionToTaskForm) validateAndSetSolution(request *AddSolutionToTaskRequest, errors map[string]response.ErrorMessage) {
	if request.Solution == "" {
		errors["solution"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.Solution = request.Solution
}
