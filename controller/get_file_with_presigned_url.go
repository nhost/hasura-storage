package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/nhost/hasura-storage/api"
	"github.com/nhost/hasura-storage/middleware"
	"github.com/sirupsen/logrus"
)

const (
	amzDateFormat = "20060102T150405Z"
)

type File struct {
	ContentType   string
	ContentLength int64
	Etag          string
	StatusCode    int
	Body          io.ReadCloser
	ExtraHeaders  http.Header
}

func expiresIn(xAmzExpires string, datestr string) (int, *APIError) {
	amzExpires, err := strconv.Atoi(xAmzExpires)
	if err != nil {
		return 0, BadDataError(err, fmt.Sprintf("problem parsing X-Amz-Expires: %s", err))
	}

	date, err := time.Parse(amzDateFormat, datestr)
	if err != nil {
		return 0, BadDataError(err, fmt.Sprintf("problem parsing X-Amz-Date: %s", err))
	}

	expires := time.Second*time.Duration(amzExpires) - time.Since(date)

	if expires <= 0 {
		return 0, BadDataError(
			errors.New("signature already expired"), //nolint: err113
			"signature already expired",
		)
	}

	return int(expires.Seconds()), nil
}

// func (ctrl *Controller) getFileWithPresignedURL(ctx *gin.Context) (*FileResponse, *APIError) {
// 	req, apiErr := ctrl.getFileWithPresignedURLParse(ctx)
// 	if apiErr != nil {
// 		return nil, apiErr
// 	}

// 	fileMetadata, _, apiErr := ctrl.getFileMetadata(
// 		ctx.Request.Context(),
// 		req.fileID,
// 		true,
// 		http.Header{"x-hasura-admin-secret": []string{ctrl.hasuraAdminSecret}},
// 	)
// 	if apiErr != nil {
// 		return nil, apiErr
// 	}

// 	downloadFunc := func() (*File, *APIError) {
// 		return ctrl.contentStorage.GetFileWithPresignedURL(
// 			ctx.Request.Context(),
// 			req.fileID,
// 			req.signature,
// 			ctx.Request.Header,
// 		)
// 	}
// 	if apiErr != nil {
// 		return nil, apiErr
// 	}

// 	return ctrl.processFileToDownload(
// 		ctx,
// 		downloadFunc,
// 		fileMetadata,
// 		fmt.Sprintf("max-age=%d", req.Expires),
// 		nil,
// 	)
// }

// func (ctrl *Controller) GetFileWithPresignedURL(ctx *gin.Context) {
// 	fileResponse, apiErr := ctrl.getFileWithPresignedURL(ctx)
// 	if apiErr != nil {
// 		_ = ctx.Error(apiErr)

// 		ctx.JSON(apiErr.statusCode, GetFileResponse{apiErr.PublicResponse()})

// 		return
// 	}

// 	defer fileResponse.body.Close()

// 	fileResponse.disableSurrageControlHeader = true
// 	fileResponse.Write(ctx)
// }

func getAmazonSignature(request api.GetPresignedURLContentsRequestObject) string {
	return fmt.Sprintf(
		"X-Amz-Algorithm=%s&X-Amz-Credential=%s&X-Amz-Date=%s&X-Amz-Expires=%s&X-Amz-Signature=%s&X-Amz-SignedHeaders=%s&X-Amz-Checksum-Mode=%s&x-id=%s", //nolint:lll
		request.Params.XAmzAlgorithm,
		request.Params.XAmzCredential,
		request.Params.XAmzDate,
		request.Params.XAmzExpires,
		request.Params.XAmzSignature,
		request.Params.XAmzSignedHeaders,
		request.Params.XAmzChecksumMode,
		request.Params.XId,
	)
}

func (ctrl *Controller) getPresignedURLContentsResponse( //nolint: ireturn,funlen,dupl
	file *processedFile,
	logger logrus.FieldLogger,
) api.GetPresignedURLContentsResponseObject {
	switch file.statusCode {
	case http.StatusOK:
		return api.GetPresignedURLContents200ApplicationoctetStreamResponse{
			Body: file.body,
			Headers: api.GetPresignedURLContents200ResponseHeaders{
				AcceptRanges: "bytes",
				CacheControl: file.cacheControl,
				ContentDisposition: fmt.Sprintf(
					`inline; filename="%s"`,
					url.QueryEscape(file.filename),
				),
				ContentType:      file.mimeType,
				Etag:             file.fileMetadata.Etag,
				LastModified:     file.fileMetadata.UpdatedAt,
				SurrogateControl: file.cacheControl,
				SurrogateKey:     file.fileMetadata.Id,
			},
			ContentLength: file.contentLength,
		}
	case http.StatusPartialContent:
		return api.GetPresignedURLContents206ApplicationoctetStreamResponse{
			Body: file.body,
			Headers: api.GetPresignedURLContents206ResponseHeaders{
				CacheControl: file.cacheControl,
				ContentDisposition: fmt.Sprintf(
					`inline; filename="%s"`,
					url.QueryEscape(file.filename),
				),
				ContentRange:     file.extraHeaders.Get("Content-Range"),
				ContentType:      file.mimeType,
				Etag:             file.fileMetadata.Etag,
				LastModified:     file.fileMetadata.UpdatedAt,
				SurrogateControl: file.cacheControl,
				SurrogateKey:     file.fileMetadata.Id,
			},
			ContentLength: file.contentLength,
		}
	case http.StatusNotModified:
		return api.GetPresignedURLContents304Response{
			Headers: api.GetPresignedURLContents304ResponseHeaders{
				CacheControl:     file.cacheControl,
				Etag:             file.fileMetadata.Etag,
				SurrogateControl: file.cacheControl,
			},
		}
	case http.StatusPreconditionFailed:
		return api.GetPresignedURLContents412Response{
			Headers: api.GetPresignedURLContents412ResponseHeaders{
				CacheControl:     file.cacheControl,
				Etag:             file.fileMetadata.Etag,
				SurrogateControl: file.cacheControl,
			},
		}
	default:
		logger.WithField("statusCode", file.statusCode).
			Error("unexpected status code from download")

		return ErrUnexpectedStatusCode
	}
}

func (ctrl *Controller) GetPresignedURLContents( //nolint: ireturn
	ctx context.Context,
	request api.GetPresignedURLContentsRequestObject,
) (api.GetPresignedURLContentsResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)
	acceptHeader := middleware.AcceptHeaderFromContext(ctx)

	fileMetadata, _, apiErr := ctrl.getFileMetadata(
		ctx,
		request.Id,
		true,
		http.Header{"x-hasura-admin-secret": []string{ctrl.hasuraAdminSecret}},
	)
	if apiErr != nil {
		logger.WithError(apiErr).Error("failed to get file metadata")
		return apiErr, nil
	}

	var httpHeaders http.Header
	if request.Params.Range != nil {
		httpHeaders = http.Header{
			"Range": []string{*request.Params.Range},
		}
	}

	expires, apiErr := expiresIn(request.Params.XAmzExpires, request.Params.XAmzDate)
	if apiErr != nil {
		logger.WithError(apiErr).Error("failed to parse expiration time")
		return apiErr, nil
	}

	fmt.Println("Signature:", getAmazonSignature(request))

	downloadFunc := func() (*File, *APIError) {
		return ctrl.contentStorage.GetFileWithPresignedURL(
			ctx,
			request.Id,
			getAmazonSignature(request),
			httpHeaders,
		)
	}

	processedFile, apiErr := ctrl.processFileToDownload(
		downloadFunc,
		fileMetadata,
		fmt.Sprintf("max-age=%d", expires),
		request.Params,
		acceptHeader,
	)
	if apiErr != nil {
		logger.WithError(apiErr).Error("failed to process file for download")
		return nil, apiErr
	}

	return ctrl.getPresignedURLContentsResponse(processedFile, logger), nil
}
