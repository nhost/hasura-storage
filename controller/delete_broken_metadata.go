package controller

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/nhost/hasura-storage/api"
)

func (ctrl *Controller) deleteBrokenMetadata(ctx *gin.Context) ([]FileSummary, *APIError) {
	missing, apiErr := ctrl.listBrokenMetadata(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	for _, m := range missing {
		if apiErr := ctrl.metadataStorage.DeleteFileByID(ctx.Request.Context(), m.ID, ctx.Request.Header); apiErr != nil {
			return nil, apiErr
		}
	}

	return missing, nil
}

func (ctrl *Controller) deleteBrokenMetadataContext(
	ctx context.Context,
) ([]FileSummary, *APIError) {
	missing, apiErr := ctrl.listBrokenMetadataContext(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	for _, m := range missing {
		if apiErr := ctrl.metadataStorage.DeleteFileByID(ctx, m.ID, nil); apiErr != nil {
			return nil, apiErr
		}
	}

	return missing, nil
}

func (ctrl *Controller) listBrokenMetadataContext(ctx context.Context) ([]FileSummary, *APIError) {
	filesInHasura, apiErr := ctrl.metadataStorage.ListFiles(ctx, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	filesInS3, apiErr := ctrl.contentStorage.ListFiles(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	missing := make([]FileSummary, 0, 10) //nolint: mnd

	for _, fileHasura := range filesInHasura {
		found := false

		for _, fileS3 := range filesInS3 {
			if path.Base(fileS3) == fileHasura.ID || !fileHasura.IsUploaded {
				found = true
			}
		}

		if !found {
			missing = append(missing, fileHasura)
		}
	}

	return missing, nil
}

func (ctrl *Controller) DeleteBrokenMetadataGin(ctx *gin.Context) {
	files, apiErr := ctrl.deleteBrokenMetadata(ctx)
	if apiErr != nil {
		_ = ctx.Error(fmt.Errorf("problem processing request: %w", apiErr))

		ctx.JSON(apiErr.statusCode, apiErr.PublicResponse())

		return
	}

	ctx.JSON(
		http.StatusOK,
		struct {
			Metadata []FileSummary `json:"metadata"`
		}{
			Metadata: files,
		},
	)
}

func (ctrl *Controller) DeleteBrokenMetadata(
	ctx context.Context,
	request api.DeleteBrokenMetadataRequestObject,
) (api.DeleteBrokenMetadataResponseObject, error) {
	files, apiErr := ctrl.deleteBrokenMetadataContext(ctx)
	if apiErr != nil {
		return apiErr, nil
	}

	// Convert from controller.FileSummary to api.FileSummary
	apiFiles := make([]api.FileSummary, len(files))
	for i, f := range files {
		apiFiles[i] = api.FileSummary{
			Id:         &f.ID,
			Name:       &f.Name,
			IsUploaded: &f.IsUploaded,
			BucketId:   &f.BucketID,
		}
	}

	return api.DeleteBrokenMetadata200JSONResponse{
		Metadata: &apiFiles,
	}, nil
}
