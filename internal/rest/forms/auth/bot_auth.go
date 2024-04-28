package auth

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/markgregr/bestHack_support_REST_server/internal/rest/forms"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
	log "github.com/sirupsen/logrus"
	"io"
)

type BotAuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type BotAuthForm struct {
	Email    string
	Password string
	Username string
}

func NewBotAuthForm() *BotAuthForm {
	return &BotAuthForm{}
}

func (f *BotAuthForm) ParseAndValidate(c *gin.Context) (forms.Former, response.Error) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()

	if err != nil {
		log.WithError(err).Error("unable to read body")
		return nil, response.NewInternalError()
	}

	var request *BotAuthRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		ve := response.NewValidationError()
		ve.SetError(response.GeneralErrorKey, response.InvalidRequestStructure, "invalid request structure")

		return nil, ve
	}

	errors := make(map[string]response.ErrorMessage)
	f.validateAndSetEmail(request, errors)
	f.validateAndSetPassword(request, errors)
	f.validateAndSetUsername(request, errors)

	if len(errors) > 0 {
		return nil, response.NewValidationError(errors)
	}

	return f, nil
}

func (f *BotAuthForm) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"email":    f.Email,
		"password": f.Password,
		"username": f.Username,
	}
}

func (f *BotAuthForm) validateAndSetEmail(request *BotAuthRequest, errors map[string]response.ErrorMessage) {
	if request.Email == "" {
		errors["email"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.Email = request.Email
}

func (f *BotAuthForm) validateAndSetPassword(request *BotAuthRequest, errors map[string]response.ErrorMessage) {
	if request.Password == "" {
		errors["password"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.Password = request.Password
}

func (f *BotAuthForm) validateAndSetUsername(request *BotAuthRequest, errors map[string]response.ErrorMessage) {
	if request.Username == "" {
		errors["username"] = response.ErrorMessage{
			Code:    response.MissedValue,
			Message: "missed value",
		}
		return
	}

	f.Username = request.Username
}
