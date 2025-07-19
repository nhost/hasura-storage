package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

const HeadersContextKey = "request.headers"

var ErrUnauthorized = errors.New("unauthorized")

func HeadersFromContext(ctx context.Context) http.Header {
	headers, _ := ctx.Value(HeadersContextKey).(http.Header)
	return headers
}

func Middleware(adminSecret string) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context,
		input *openapi3filter.AuthenticationInput,
	) error {
		if input.SecuritySchemeName == "X-Hasura-Admin-Secret" {
			adminSecretHeader := input.RequestValidationInput.Request.Header.Get(
				"X-Hasura-Admin-Secret",
			)
			if adminSecretHeader != adminSecret {
				return ErrUnauthorized
			}
		}

		c := ginmiddleware.GetGinContext(ctx)
		c.Set(HeadersContextKey, input.RequestValidationInput.Request.Header)
		return nil
	}
}
