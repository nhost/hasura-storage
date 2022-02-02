package cmd

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/gin-gonic/gin"
	"github.com/nhost/hasura-storage/controller"
	"github.com/nhost/hasura-storage/metadata"
	"github.com/nhost/hasura-storage/migrations"
	"github.com/nhost/hasura-storage/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	bind                      = "bind"
	graphqlEndpoint           = "graphql_endpoint"
	s3Endpoint                = "s3_endpoint"
	s3AccessKey               = "s3_access_key"
	s3SecretKey               = "s3_secret_key" // nolint: gosec
	s3Region                  = "s3_region"
	s3Bucket                  = "s3_bucket"
	s3RootFolder              = "s3_root_folder"
	postgresMigrations        = "postgres-migrations"
	postgresMigrationsSource  = "postgres-migrations-source"
	hasuraMetadata            = "hasura-metadata"
	hasuraMetadataAdminSecret = "hasura-metadata-admin-secret" // nolint: gosec
)

func ginLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()

		ctx.Next()

		endTime := time.Now()

		latencyTime := endTime.Sub(startTime)
		reqMethod := ctx.Request.Method
		reqURL := ctx.Request.RequestURI
		statusCode := ctx.Writer.Status()
		clientIP := ctx.ClientIP()

		fields := logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"method":       reqMethod,
			"url":          reqURL,
			"errors":       ctx.Errors.Errors(),
		}

		if len(ctx.Errors.Errors()) > 0 {
			logger.WithFields(fields).Error("call completed with some errors")
		} else {
			logger.WithFields(fields).Info()
		}
	}
}

func getGin(
	metadataStorage controller.MetadataStorage,
	contentStorage controller.ContentStorage,
	logger *logrus.Logger,
	debug bool,
) *gin.Engine {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	ctrl := controller.New(metadataStorage, contentStorage, logger)

	return ctrl.SetupRouter(ginLogger(logger))
}

func getMetadataStorage(endpoint string) *metadata.Hasura {
	return metadata.NewHasura(endpoint, metadata.ForWardHeadersAuthorizer)
}

func getContentStorage(
	s3Endpoint, region, s3AccessKey, s3SecretKey, bucket, rootFolder string, logger *logrus.Logger,
) *storage.S3 {
	config := &aws.Config{ // nolint: exhaustivestruct
		Credentials:      credentials.NewStaticCredentials(s3AccessKey, s3SecretKey, ""),
		Endpoint:         aws.String(s3Endpoint),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}

	st, err := storage.NewS3(config, bucket, rootFolder, logger)
	if err != nil {
		panic(err)
	}

	return st
}

func applymigrations(
	postgresMigrations bool,
	postgresSource string,
	hasuraMetadata bool,
	hasruraEndpoint string,
	hasuraSecret string,
	logger *logrus.Logger,
) {
	if postgresMigrations {
		if postgresSource == "" {
			logger.Error("you need to specify --postgres-migrations-source")
			os.Exit(1)
		}
		logger.Info("applying postgres migrations")
		if err := migrations.ApplyPostgresMigration(postgresSource); err != nil {
			logger.Errorf("problem applying postgres migrations: %s", err.Error())
			os.Exit(1)
		}
	}

	if hasuraMetadata {
		logger.Info("applying hasura metadata")
		if err := migrations.ApplyHasuraMetadata(hasruraEndpoint, hasuraSecret, logger); err != nil {
			logger.Errorf("problem applying hasura metadata: %s", err.Error())
			os.Exit(1)
		}
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)

	addStringFlag(serveCmd.Flags(), bind, ":8000", "bind the service to this address")

	{
		addStringFlag(
			serveCmd.Flags(),
			graphqlEndpoint,
			"",
			"Use this endpoint when connecting using graphql as metadata storage",
		)
	}

	{
		addStringFlag(serveCmd.Flags(), s3Endpoint, "", "S3 Endpoint")
		addStringFlag(serveCmd.Flags(), s3AccessKey, "", "S3 Access key")
		addStringFlag(serveCmd.Flags(), s3SecretKey, "", "S3 Secret key")
		addStringFlag(serveCmd.Flags(), s3Region, "", "S3 region")
		addStringFlag(serveCmd.Flags(), s3Bucket, "", "S3 bucket")
		addStringFlag(serveCmd.Flags(), s3RootFolder, "", "All buckets will be created inside this root")
	}

	{
		addBoolFlag(serveCmd.Flags(), postgresMigrations, false, "Apply Postgres migrations")
		addStringFlag(
			serveCmd.Flags(),
			postgresMigrationsSource,
			"",
			"postgres connection, i.e. postgres://user@pass:localhost:5432/mydb",
		)
	}

	{
		addBoolFlag(serveCmd.Flags(), hasuraMetadata, false, "Apply Hasura's metadata")
		addStringFlag(serveCmd.Flags(), hasuraMetadataAdminSecret, "", "")
	}
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts hasura-storage server",
	Run: func(_ *cobra.Command, _ []string) {
		logger := getLogger()

		logger.Info("storage version ", controller.Version())

		if viper.GetBool(debug) {
			logger.SetLevel(logrus.DebugLevel)
			gin.SetMode(gin.DebugMode)
		} else {
			logger.SetLevel(logrus.InfoLevel)
			gin.SetMode(gin.ReleaseMode)
		}

		logger.WithFields(
			logrus.Fields{
				"debug":               viper.GetBool(debug),
				"bind":                viper.GetString(bind),
				"graphql_endpoint":    viper.GetString(graphqlEndpoint),
				"postgres-migrations": viper.GetBool(postgresMigrations),
				"hasura-metadata":     viper.GetBool(hasuraMetadata),
				"s3_endpoint":         viper.GetString(s3Endpoint),
				"s3_region":           viper.GetString(s3Region),
				"s3_bucket":           viper.GetString(s3Bucket),
				"s3_root_folder":      viper.GetString(s3RootFolder),
			},
		).Debug("parameters")

		contentStorage := getContentStorage(
			viper.GetString(s3Endpoint),
			viper.GetString(s3Region),
			viper.GetString(s3AccessKey),
			viper.GetString(s3SecretKey),
			viper.GetString(s3Bucket),
			viper.GetString(s3RootFolder),
			logger,
		)

		applymigrations(
			viper.GetBool(postgresMigrations),
			viper.GetString(postgresMigrationsSource),
			viper.GetBool(hasuraMetadata),
			viper.GetString(graphqlEndpoint),
			viper.GetString(hasuraMetadataAdminSecret),
			logger,
		)

		metadataStorage := getMetadataStorage(
			viper.GetString(graphqlEndpoint) + "/graphql",
		)
		r := getGin(metadataStorage, contentStorage, logger, viper.GetBool(debug))

		logger.Info("starting server")
		logger.Error(r.Run(viper.GetString(bind)))
	},
}
