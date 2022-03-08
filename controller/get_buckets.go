package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// this type is used to ensure we respond consistently no matter the case.
type GetBucketsResponse struct {
	Buckets []BucketMetadata `json:"buckets,omitempty"`
	Error   *ErrorResponse   `json:"error,omitempty"`
}

func (ctrl *Controller) getBuckets(ctx *gin.Context) ([]BucketMetadata, *APIError) {
	return ctrl.metadataStorage.GetBuckets(ctx.Request.Context(), ctx.Request.Header)
}

func (ctrl *Controller) GetBuckets(ctx *gin.Context) {
	bucketsMetadata, apiErr := ctrl.getBuckets(ctx)
	if apiErr != nil {
		_ = ctx.Error(fmt.Errorf("problem processing request: %w", apiErr))

		ctx.JSON(apiErr.statusCode, GetBucketsResponse{nil, apiErr.PublicResponse()})

		return
	}

	ctx.JSON(http.StatusCreated, GetBucketsResponse{bucketsMetadata, nil})
}
