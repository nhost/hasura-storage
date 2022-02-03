package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nhost/hasura-storage/controller"
	"github.com/nhost/hasura-storage/controller/mock_controller"
	"github.com/sirupsen/logrus"
)

func getFileTestCases() []struct {
	name           string
	requestHeaders http.Header
	expectedStatus int
} {
	cases := []struct {
		name           string
		requestHeaders http.Header
		expectedStatus int
	}{
		{
			name:           "no headers",
			expectedStatus: 200,
			requestHeaders: http.Header{},
		},
		{
			name:           "if-match matches",
			expectedStatus: 200,
			requestHeaders: http.Header{
				"if-match": {"\"55af1e60-0f28-454e-885e-ea6aab2bb288\""},
			},
		},
		{
			name:           "if-match doesn't match",
			expectedStatus: 412,
			requestHeaders: http.Header{
				"if-match": {"blah"},
			},
		},
		{
			name:           "if-none-match matches",
			expectedStatus: 304,
			requestHeaders: http.Header{
				"if-none-match": {"\"55af1e60-0f28-454e-885e-ea6aab2bb288\""},
			},
		},
		{
			name:           "if-none-match doesn't match",
			expectedStatus: 200,
			requestHeaders: http.Header{
				"if-none-match": {"blah"},
			},
		},
		{
			name:           "if-modified since matches",
			expectedStatus: 200,
			requestHeaders: http.Header{
				"if-modified-since": {"Wed, 15 Jan 2020 10:00:00 UTC"},
			},
		},
		{
			name:           "if-modified doesn't match",
			expectedStatus: 304,
			requestHeaders: http.Header{
				"if-modified-since": {"Thu, 25 Jan 2024 10:00:00 UTC"},
			},
		},
		{
			name:           "if-unmodified-since matches",
			expectedStatus: 200,
			requestHeaders: http.Header{
				"if-unmodified-since": {"Thu, 25 Jan 2024 10:00:00 UTC"},
			},
		},
		{
			name:           "if-unmodified-since doesn't match",
			expectedStatus: 412,
			requestHeaders: http.Header{
				"if-unmodified-since": {"Wed, 15 Jan 2020 10:00:00 UTC"},
			},
		},
	}
	return cases
}

func TestGetFileInfo(t *testing.T) {
	t.Parallel()

	cases := getFileTestCases()

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

			metadataStorage.EXPECT().GetFileByID(
				gomock.Any(), "55af1e60-0f28-454e-885e-ea6aab2bb288", gomock.Any(),
			).Return(controller.FileMetadataWithBucket{
				FileMetadata: controller.FileMetadata{
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
				},
				Bucket: controller.BucketMetadata{
					ID:                   "default",
					MinUploadFile:        0,
					MaxUploadFile:        100,
					PresignedURLsEnabled: true,
					DownloadExpiration:   30,
					CreatedAt:            "2021-12-15T13:26:52.082485+00:00",
					UpdatedAt:            "2021-12-15T13:26:52.082485+00:00",
					CacheControl:         "max-age=3600",
				},
			}, nil)

			ctrl := controller.New(metadataStorage, contentStorage, logger)

			router, _ := ctrl.SetupRouter(nil, ginLogger(logger))

			responseRecorder := httptest.NewRecorder()

			req, _ := http.NewRequestWithContext(
				context.Background(),
				"HEAD",
				"/v1/storage/files/55af1e60-0f28-454e-885e-ea6aab2bb288",
				nil,
			)

			req.Header.Add("x-hasura-user-id", "some-valid-uuid")
			for k, v := range tc.requestHeaders {
				for _, vv := range v {
					req.Header.Add(k, vv)
				}
			}

			router.ServeHTTP(responseRecorder, req)

			assert(t, tc.expectedStatus, responseRecorder.Code)

			assert(t, responseRecorder.Header(), http.Header{
				"Cache-Control":  {"max-age=3600"},
				"Content-Length": {"64"},
				"Content-Type":   {"text/plain; charset=utf-8"},
				"Etag":           {`"55af1e60-0f28-454e-885e-ea6aab2bb288"`},
				"Last-Modified":  {"Mon, 27 Dec 2021 09:58:11 UTC"},
			})
		})
	}
}
