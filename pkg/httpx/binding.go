// Package httpx holds small HTTP helpers shared across feature packages, such
// as turning gin binding failures into structured, field-level error responses.
package httpx

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// FieldError describes a single failed validation constraint on a request field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// UseJSONFieldNames makes the validator report the json tag name (e.g. "email")
// instead of the Go struct field name ("Email") in validation errors. Call it
// once at startup, before the server begins handling requests.
func UseJSONFieldNames() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidationErrors converts a gin binding error into a field-level list. It
// returns (nil, false) when err is not a validation error (for example a
// malformed JSON body), so callers can fall back to a generic message.
func ValidationErrors(err error) ([]FieldError, bool) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil, false
	}

	out := make([]FieldError, 0, len(ve))
	for _, fe := range ve {
		out = append(out, FieldError{
			Field:   fe.Field(),
			Message: messageForTag(fe),
		})
	}
	return out, true
}

// WriteBindError writes a 400 response for a binding failure: a field-level
// list when the error is a validation error, otherwise a generic message for
// malformed input. Callers should return immediately after invoking it.
func WriteBindError(c *gin.Context, err error) {
	if fields, ok := ValidationErrors(err); ok {
		WriteValidationError(c, fields)
		return
	}
	WriteError(c, http.StatusBadRequest, "invalid request payload")
}

// messageForTag renders a human-readable message for a failed constraint.
func messageForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	case "uuid":
		return "must be a valid uuid"
	default:
		return "invalid value"
	}
}
