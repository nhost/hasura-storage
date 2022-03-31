package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetFileWithPresignedURLRequest struct {
	fileID    string
	signature string
	headers   getFileInformationHeaders
}

func (ctrl *Controller) getFileWithPresignedURLParse(ctx *gin.Context) (GetFileWithPresignedURLRequest, *APIError) {
	var headers getFileInformationHeaders
	if err := ctx.ShouldBindHeader(&headers); err != nil {
		return GetFileWithPresignedURLRequest{}, InternalServerError(fmt.Errorf("problem parsing request headers: %w", err))
	}

	return GetFileWithPresignedURLRequest{
		fileID:    ctx.Param("id"),
		signature: ctx.Request.URL.RawQuery,
		headers:   headers,
	}, nil
}

func (ctrl *Controller) getFileWithPresignedURL(ctx *gin.Context) (int, *APIError) {
	req, apiErr := ctrl.getFileWithPresignedURLParse(ctx)
	if apiErr != nil {
		return 0, apiErr
	}

	fileMetadata, apiErr := ctrl.getFileMetadata(
		ctx.Request.Context(),
		req.fileID,
		http.Header{"x-hasura-admin-secret": []string{ctrl.hasuraAdminSecret}},
	)
	if apiErr != nil {
		return 0, apiErr
	}

	object, apiErr := ctrl.contentStorage.GetFileWithPresignedURL(
		ctx.Request.Context(),
		req.fileID,
		req.signature,
		ctx.Request.Header,
	)
	if apiErr != nil {
		return 0, apiErr
	}
	defer object.Close()

	return ctrl.processFileToDownload(ctx, object, fileMetadata, req.headers)
}

func (ctrl *Controller) GetFileWithPresignedURL(ctx *gin.Context) {
	statusCode, apiErr := ctrl.getFileWithPresignedURL(ctx)

	if apiErr != nil {
		_ = ctx.Error(apiErr)

		ctx.JSON(apiErr.statusCode, GetFileResponse{apiErr.PublicResponse()})

		return
	}

	ctx.AbortWithStatus(statusCode)
}
