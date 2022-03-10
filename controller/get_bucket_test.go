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

func TestGetBucket(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		expectedStatus   int
		expectedResponse *controller.GetBucketResponse
	}{
		{
			name:           "good",
			expectedStatus: 200,
			expectedResponse: &controller.GetBucketResponse{
				BucketMetadata: &controller.BucketMetadata{
					ID:                   "blah",
					MinUploadFile:        0,
					MaxUploadFile:        100,
					PresignedURLsEnabled: true,
					DownloadExpiration:   30,
					CreatedAt:            "2021-12-15T13:26:52.082485+00:00",
					UpdatedAt:            "2021-12-15T13:26:52.082485+00:00",
					CacheControl:         "",
				},
			},
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

			metadataStorage.EXPECT().GetBucketByID(
				gomock.Any(), "blah", gomock.Any(),
			).Return(controller.BucketMetadata{
				ID:                   "blah",
				MinUploadFile:        0,
				MaxUploadFile:        100,
				PresignedURLsEnabled: true,
				DownloadExpiration:   30,
				CreatedAt:            "2021-12-15T13:26:52.082485+00:00",
				UpdatedAt:            "2021-12-15T13:26:52.082485+00:00",
			}, nil)

			ctrl := controller.New(metadataStorage, contentStorage, logger)

			router, _ := ctrl.SetupRouter(nil, ginLogger(logger))

			responseRecorder := httptest.NewRecorder()

			req, _ := http.NewRequestWithContext(
				context.Background(),
				"GET",
				"/v1/storage/buckets/blah",
				nil,
			)

			router.ServeHTTP(responseRecorder, req)

			assert(t, tc.expectedStatus, responseRecorder.Code)
			assertJSON(t, responseRecorder.Body.Bytes(), &controller.GetBucketResponse{}, tc.expectedResponse)
		})
	}
}
