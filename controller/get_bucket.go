package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// this type is used to ensure we respond consistently no matter the case.
type GetBucketResponse struct {
	*BucketMetadata
	Error *ErrorResponse `json:"error,omitempty"`
}

func (ctrl *Controller) getBucket(ctx *gin.Context) (BucketMetadata, *APIError) {
	id := ctx.Param("id")
	return ctrl.metadataStorage.GetBucketByID(ctx.Request.Context(), id, ctx.Request.Header)
}

func (ctrl *Controller) GetBucket(ctx *gin.Context) {
	bucketMetadata, apiErr := ctrl.getBucket(ctx)
	if apiErr != nil {
		_ = ctx.Error(fmt.Errorf("problem processing request: %w", apiErr))

		ctx.JSON(apiErr.statusCode, GetBucketResponse{nil, apiErr.PublicResponse()})

		return
	}

	ctx.JSON(http.StatusOK, GetBucketResponse{&bucketMetadata, nil})
}
