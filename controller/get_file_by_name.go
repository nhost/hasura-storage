package controller

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) getFileMetadataByName(
	ctx context.Context,
	bucket string,
	name string,
	headers http.Header,
) (FileMetadataWithBucket, *APIError) {
	fileMetadata, err := ctrl.metadataStorage.GetFileByName(ctx, bucket, name, headers)
	if err != nil {
		return FileMetadataWithBucket{}, err
	}

	if !fileMetadata.IsUploaded {
		return FileMetadataWithBucket{}, ForbiddenError(nil, "you are not auhtorized")
	}

	return fileMetadata, nil
}

func (ctrl *Controller) getFileByNameProcess(ctx *gin.Context) (int, *APIError) {
	req, apiErr := ctrl.getFileParse(ctx)
	if apiErr != nil {
		return 0, apiErr
	}

	filename := ctx.Query("filename")
	if filename == "" {
		return 0, ErrMissingFilename
	}

	bucketID := ctx.Param("id")
	fileMetadata, apiErr := ctrl.getFileMetadataByName(ctx.Request.Context(), bucketID, filename, ctx.Request.Header)
	if apiErr != nil {
		return 0, apiErr
	}

	return ctrl.getFile(ctx, fileMetadata, req.headers)
}

func (ctrl *Controller) GetFileByName(ctx *gin.Context) {
	statusCode, apiErr := ctrl.getFileByNameProcess(ctx)
	if apiErr != nil {
		_ = ctx.Error(apiErr)

		ctx.JSON(apiErr.statusCode, GetFileResponse{apiErr.PublicResponse()})

		return
	}

	ctx.AbortWithStatus(statusCode)
}
