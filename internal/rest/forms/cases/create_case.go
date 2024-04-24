package cases

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/markgregr/bestHack_support_REST_server/internal/rest/forms"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
	log "github.com/sirupsen/logrus"
	"io"
)

type CreateCaseRequest struct {
	Title    string `json:"title"`
	Solution string `json:"solution"`
}

type CreateCaseForm struct {
	Title    string
	Solution string
}

func NewCreateCaseForm() *CreateCaseForm {
	return &CreateCaseForm{}
}

func (f *CreateCaseForm) ParseAndValidate(c *gin.Context) (forms.Former, response.Error) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()

	if err != nil {
		log.WithError(err).Error("unable to read body")
		return nil, response.NewInternalError()
	}

	var request *CreateCaseRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}
	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetTitle(request, errors)
	f.validateAndSetSolution(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *CreateCaseForm) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"title":    f.Title,
		"solution": f.Solution,
	}
}

func (f *CreateCaseForm) validateAndSetTitle(request *CreateCaseRequest, errors map[string]response.ErrorMessage) {
	if request.Title == "" {
		errors["title"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.Title = request.Title
}

func (f *CreateCaseForm) validateAndSetSolution(request *CreateCaseRequest, errors map[string]response.ErrorMessage) {
	if request.Solution == "" {
		errors["solution"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.Solution = request.Solution
}
