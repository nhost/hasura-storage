package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) deleteFile(ctx *gin.Context) *APIError {
	id := ctx.Param("id")

	apiErr := ctrl.metadataStorage.DeleteFileByID(ctx.Request.Context(), id, ctx.Request.Header)
	if apiErr != nil {
		return apiErr
	}

	if apiErr := ctrl.contentStorage.DeleteFile(id); apiErr != nil {
		return apiErr
	}

	ctx.Set("FileChanged", id)

	return nil
}

func (ctrl *Controller) DeleteFile(ctx *gin.Context) {
	apiErr := ctrl.deleteFile(ctx)
	if apiErr != nil {
		_ = ctx.Error(fmt.Errorf("problem processing request: %w", apiErr))

		ctx.JSON(apiErr.statusCode, apiErr.PublicResponse())

		return
	}

	ctx.AbortWithStatus(http.StatusNoContent)
}
