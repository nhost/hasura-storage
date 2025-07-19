package controller

import (
	"bytes"
	"context"
	_ "embed"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nhost/hasura-storage/api"
)

//go:embed openapi.yaml
var OpenAPISchema []byte

func (ctrl *Controller) OpenAPI(ctx *gin.Context) {
	content := bytes.NewReader(OpenAPISchema)
	http.ServeContent(
		ctx.Writer,
		ctx.Request,
		"openapi.yaml",
		time.Now(), // we should inject this at compile time
		content,
	)
}

func (ctrl *Controller) GetOpenAPISpec(
	ctx context.Context,
	request api.GetOpenAPISpecRequestObject,
) (api.GetOpenAPISpecResponseObject, error) {
	return nil, nil
}
