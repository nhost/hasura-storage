//go:generate mockgen -destination mock/controller.go -package mock -source=controller.go MetadataStorage
package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nhost/hasura-storage/api"
	"github.com/nhost/hasura-storage/image"
	"github.com/sirupsen/logrus"
)

type FileSummary struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsUploaded bool   `json:"isUploaded"`
	BucketID   string `json:"bucketId"`
}

type BucketMetadata struct {
	ID                   string
	MinUploadFile        int
	MaxUploadFile        int
	PresignedURLsEnabled bool
	DownloadExpiration   int
	CreatedAt            string
	UpdatedAt            string
	CacheControl         string
}

type MetadataStorage interface {
	GetBucketByID(ctx context.Context, id string, headers http.Header) (BucketMetadata, *APIError)
	GetFileByID(ctx context.Context, id string, headers http.Header) (api.FileMetadata, *APIError)
	InitializeFile(
		ctx context.Context,
		id, name string, size int64, bucketID, mimeType string,
		headers http.Header,
	) *APIError
	PopulateMetadata(
		ctx context.Context,
		id, name string, size int64, bucketID, etag string, IsUploaded bool, mimeType string,
		metadata map[string]any,
		headers http.Header) (api.FileMetadata, *APIError,
	)
	SetIsUploaded(
		ctx context.Context,
		fileID string,
		isUploaded bool,
		headers http.Header,
	) *APIError
	DeleteFileByID(ctx context.Context, fileID string, headers http.Header) *APIError
	ListFiles(ctx context.Context, headers http.Header) ([]FileSummary, *APIError)
	InsertVirus(
		ctx context.Context,
		fileID, filename, virus string,
		userSession map[string]any,
		headers http.Header,
	) *APIError
}

type ContentStorage interface {
	PutFile(
		ctx context.Context,
		content io.ReadSeeker,
		filepath, contentType string,
	) (string, *APIError)
	GetFile(ctx context.Context, filepath string, downloadRange *string) (*File, *APIError)
	CreatePresignedURL(
		ctx context.Context,
		filepath string,
		expire time.Duration,
	) (string, *APIError)
	GetFileWithPresignedURL(
		ctx context.Context, filepath, signature string, headers http.Header,
	) (*File, *APIError)
	DeleteFile(ctx context.Context, filepath string) *APIError
	ListFiles(ctx context.Context) ([]string, *APIError)
}

type Antivirus interface {
	ScanReader(r io.ReaderAt) *APIError
}

type Controller struct {
	publicURL         string
	apiRootPrefix     string
	hasuraAdminSecret string
	metadataStorage   MetadataStorage
	contentStorage    ContentStorage
	imageTransformer  *image.Transformer
	av                Antivirus
	logger            *logrus.Logger
}

func New(
	publicURL string,
	apiRootPrefix string,
	hasuraAdminSecret string,
	metadataStorage MetadataStorage,
	contentStorage ContentStorage,
	imageTransformer *image.Transformer,
	av Antivirus,
	logger *logrus.Logger,
) *Controller {
	return &Controller{
		publicURL,
		apiRootPrefix,
		hasuraAdminSecret,
		metadataStorage,
		contentStorage,
		imageTransformer,
		av,
		logger,
	}
}

func corsConfig(allowedOrigins []string) cors.Config {
	return cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "PUT", "POST", "HEAD", "DELETE"},
		AllowHeaders: []string{
			"Authorization", "Origin", "if-match", "if-none-match", "if-modified-since", "if-unmodified-since",
			"x-hasura-admin-secret", "x-nhost-bucket-id", "x-nhost-file-name", "x-nhost-file-id",
			"x-hasura-role",
		},
		ExposeHeaders: []string{
			"Content-Length", "Content-Type", "Cache-Control", "ETag", "Last-Modified", "X-Error",
		},
		MaxAge: 12 * time.Hour, //nolint: mnd
	}
}

func (ctrl *Controller) SetupRouter(
	trustedProxies []string,
	apiRootPrefix string,
	corsOrigins []string,
	corsAllowCredentials bool,
	middleware ...gin.HandlerFunc,
) (*gin.Engine, error) {
	router := gin.New()
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		return nil, fmt.Errorf("problem setting trusted proxies: %w", err)
	}

	// lower values make uploads slower but keeps service memory usage low
	router.MaxMultipartMemory = 1 << 20 //nolint:mnd  // 1 MB
	router.Use(gin.Recovery())

	for _, mw := range middleware {
		router.Use(mw)
	}

	corsConfig := corsConfig(corsOrigins)
	if corsAllowCredentials {
		corsConfig.AllowCredentials = true
	}

	router.Use(cors.New(corsConfig))

	router.GET("/healthz", ctrl.Health)

	apiRoot := router.Group(apiRootPrefix)
	{
		apiRoot.GET("/openapi.yaml", ctrl.OpenAPI)
		apiRoot.GET("/version", ctrl.Version)
	}
	files := apiRoot.Group("/files")
	{
		// files.POST("", ctrl.UploadFilesGin) // To delete
		// files.POST("/", ctrl.UploadFilesGin)
		// files.GET("/:id", ctrl.GetFileGin)
		// files.HEAD("/:id", ctrl.GetFileInformation)
		files.PUT("/:id", ctrl.ReplaceFileGin)
		files.DELETE("/:id", ctrl.DeleteFileGin)
		files.GET("/:id/presignedurl", ctrl.GetFilePresignedURL)
		// files.GET("/:id/presignedurl/content", ctrl.GetFileWithPresignedURL)
	}

	ops := apiRoot.Group("/ops")
	{
		ops.POST("list-orphans", ctrl.ListOrphanedFilesGin)
		ops.POST("delete-orphans", ctrl.DeleteOrphans)
		ops.POST("list-broken-metadata", ctrl.ListBrokenMetadataGin)
		ops.POST("delete-broken-metadata", ctrl.DeleteBrokenMetadataGin)
		ops.POST("list-not-uploaded", ctrl.ListNotUploaded)
	}
	return router, nil
}

func (ctrl *Controller) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"healthz": "ok",
	})
}
