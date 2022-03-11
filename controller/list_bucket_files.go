package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// this type is used to ensure we respond consistently no matter the case.
type ListBucketFilesResponse struct {
	Files []FileMetadata `json:"files,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`
}

func (ctrl *Controller) listBucketFiles(ctx *gin.Context) ([]FileMetadata, *APIError) {
	id := ctx.Param("id")
	filter := ctx.Query("filter_files_regex")
	return ctrl.metadataStorage.ListBucketFiles(ctx.Request.Context(), id, filter, ctx.Request.Header)
}

func (ctrl *Controller) ListBucketFiles(ctx *gin.Context) {
	bucketFiles, apiErr := ctrl.listBucketFiles(ctx)
	if apiErr != nil {
		_ = ctx.Error(fmt.Errorf("problem processing request: %w", apiErr))

		ctx.JSON(apiErr.statusCode, ListBucketFilesResponse{nil, apiErr.PublicResponse()})

		return
	}

	ctx.JSON(http.StatusOK, ListBucketFilesResponse{bucketFiles, nil})
}
