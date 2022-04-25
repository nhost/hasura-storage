package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nhost/hasura-storage/controller"
	"github.com/nhost/hasura-storage/controller/mock_controller"
	"github.com/sirupsen/logrus"
)

func TestGetFilePresignedURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		expectedStatus int
		fileFound      bool
	}{
		{
			name:           "success",
			expectedStatus: 200,
			fileFound:      true,
		},
		{
			name:           "fileNotFound",
			expectedStatus: 404,
			fileFound:      false,
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := gomock.NewController(t)
			defer c.Finish()

			metadataStorage := mock_controller.NewMockMetadataStorage(c)
			contentStorage := mock_controller.NewMockContentStorage(c)

			if tc.fileFound {
				metadataStorage.EXPECT().GetFileByID(
					gomock.Any(), "55af1e60-0f28-454e-885e-ea6aab2bb288", gomock.Any(),
				).Return(controller.FileMetadata{
					ID:               "55af1e60-0f28-454e-885e-ea6aab2bb288",
					Name:             "my-file.txt",
					Size:             64,
					BucketID:         "default",
					ETag:             "\"55af1e60-0f28-454e-885e-ea6aab2bb288\"",
					CreatedAt:        "2021-12-27T09:58:11Z",
					UpdatedAt:        "2021-12-27T09:58:11Z",
					IsUploaded:       true,
					MimeType:         "text/plain; charset=utf-8",
					UploadedByUserID: "0f7f0ff0-f945-4597-89e1-3636b16775cd",
				}, nil)

				metadataStorage.EXPECT().GetBucketByID(
					gomock.Any(), "default", gomock.Any(),
				).Return(controller.BucketMetadata{
					ID:                   "default",
					MinUploadFile:        0,
					MaxUploadFile:        100,
					PresignedURLsEnabled: true,
					DownloadExpiration:   30,
					CreatedAt:            "2021-12-15T13:26:52.082485+00:00",
					UpdatedAt:            "2021-12-15T13:26:52.082485+00:00",
					CacheControl:         "max-age=3600",
				}, nil)

				contentStorage.EXPECT().CreatePresignedURL(
					"55af1e60-0f28-454e-885e-ea6aab2bb288", 30*time.Minute,
				).Return(
					"this-is-the-signature", nil,
				)
			} else {
				metadataStorage.EXPECT().GetFileByID(
					gomock.Any(), "55af1e60-0f28-454e-885e-ea6aab2bb288", gomock.Any(),
				).Return(controller.FileMetadata{},
					controller.ErrFileNotFound)
			}

			ctrl := controller.New("http://asd", "asdasd", metadataStorage, contentStorage, nil, logger)

			router, _ := ctrl.SetupRouter(nil, ginLogger(logger))

			responseRecorder := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(
				context.Background(),
				"GET",
				"/v1/storage/files/55af1e60-0f28-454e-885e-ea6aab2bb288/presignedurl",
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			router.ServeHTTP(responseRecorder, req)

			assert(t, tc.expectedStatus, responseRecorder.Code)

			assert(t, responseRecorder.Header(), http.Header{
				"Content-Type": {"application/json; charset=utf-8"},
			})

			resp := &controller.GetFilePresignedURLResponse{}
			if err := json.Unmarshal(responseRecorder.Body.Bytes(), resp); err != nil {
				t.Error(err)
			}

			if tc.fileFound {
				assert(t, resp, &controller.GetFilePresignedURLResponse{
					URL:        "http://asd/v1/storage/files/55af1e60-0f28-454e-885e-ea6aab2bb288/presignedurl/content?this-is-the-signature", // nolint: lll
					Expiration: 30,
				})
			} else {
				assert(t, resp, &controller.GetFilePresignedURLResponse{
					Error: &controller.ErrorResponse{Message: "file not found"},
				})
			}
		})
	}
}
