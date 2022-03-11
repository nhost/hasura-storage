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

func TestGetBucketFiles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		expectedStatus   int
		expectedResponse *controller.GetBucketFilesResponse
	}{
		{
			name:           "good",
			expectedStatus: 200,
			expectedResponse: &controller.GetBucketFilesResponse{
				Files: []controller.FileMetadata{
					{
						ID:               "38288c85-02af-416b-b075-11c4dae9",
						Name:             "a_file.txt",
						Size:             12,
						BucketID:         "blah",
						ETag:             "some-etag",
						CreatedAt:        "",
						UpdatedAt:        "",
						IsUploaded:       true,
						MimeType:         "text/plain; charset=utf-8",
						UploadedByUserID: "some-valid-uuid",
					},
					{
						ID:               "d041c7c5-10e7-410e-a599-799409b5",
						Name:             "another_file.md",
						Size:             12,
						BucketID:         "blah",
						ETag:             "some-etag",
						CreatedAt:        "",
						UpdatedAt:        "",
						IsUploaded:       true,
						MimeType:         "text/plain; charset=utf-8",
						UploadedByUserID: "some-valid-uuid",
					},
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

			metadataStorage.EXPECT().GetBucketFiles(
				gomock.Any(), "blah", "", gomock.Any(),
			).Return([]controller.FileMetadata{
				{
					ID:               "38288c85-02af-416b-b075-11c4dae9",
					Name:             "a_file.txt",
					Size:             12,
					BucketID:         "blah",
					ETag:             "some-etag",
					CreatedAt:        "",
					UpdatedAt:        "",
					IsUploaded:       true,
					MimeType:         "text/plain; charset=utf-8",
					UploadedByUserID: "some-valid-uuid",
				},
				{
					ID:               "d041c7c5-10e7-410e-a599-799409b5",
					Name:             "another_file.md",
					Size:             12,
					BucketID:         "blah",
					ETag:             "some-etag",
					CreatedAt:        "",
					UpdatedAt:        "",
					IsUploaded:       true,
					MimeType:         "text/plain; charset=utf-8",
					UploadedByUserID: "some-valid-uuid",
				},
			}, nil)

			ctrl := controller.New(metadataStorage, contentStorage, logger)

			router, _ := ctrl.SetupRouter(nil, ginLogger(logger))

			responseRecorder := httptest.NewRecorder()

			req, _ := http.NewRequestWithContext(
				context.Background(),
				"GET",
				"/v1/storage/buckets/blah/files/list",
				nil,
			)

			router.ServeHTTP(responseRecorder, req)

			assert(t, tc.expectedStatus, responseRecorder.Code)
			assertJSON(t, responseRecorder.Body.Bytes(), &controller.GetBucketFilesResponse{}, tc.expectedResponse)
		})
	}
}
