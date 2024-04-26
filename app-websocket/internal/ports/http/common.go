package http

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
)

type ErrorBody struct {
	Message string `json:"message"`
}

func ProcessError(w http.ResponseWriter, msg string, code int) {
	body := ErrorBody{
		Message: msg,
	}
	buf, _ := json.Marshal(body)

	w.WriteHeader(code)
	_, _ = w.Write(buf)
}

func ValidationError(errs validator.ValidationErrors) string {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return strings.Join(errMsgs, ", ")
}
