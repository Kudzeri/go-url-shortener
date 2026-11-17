package response

import "github.com/go-playground/validator/v10"

type Response struct {
	Status string            `json:"status"`
	Error  string            `json:"error,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() *Response {
	return &Response{
		Status: StatusOK,
	}
}

func Error(msg string) *Response {
	return &Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) *Response {
	return &Response{
		Status: StatusError,
		Error:  "Ошибка валидации",
		Fields: ValidationErrors(errs),
	}
}

func ValidationErrors(errs validator.ValidationErrors) map[string]string {
	fields := make(map[string]string)

	for _, err := range errs {
		switch err.Field() {
		case "URL":
			switch err.Tag() {
			case "required":
				fields["url"] = "Поле URL является обязательным"
			case "http_url":
				fields["url"] = "Укажите корректный URL"
			}
		case "Alias":
			switch err.Tag() {
			case "min", "max":
				fields["alias"] = "Alias должен содержать от 3 до 32 символов"
			case "alphanum":
				fields["alias"] = "Alias может содержать латинские буквы и цифры"
			}
		}
	}

	return fields
}
