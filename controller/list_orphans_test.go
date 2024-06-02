package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nhost/hasura-storage/controller"
	"github.com/nhost/hasura-storage/controller/mock"
	"github.com/sirupsen/logrus"
	gomock "go.uber.org/mock/gomock"
)

func TestListOrphans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		expected controller.ListOrphansResponse
	}{
		{
			name: "successful",
			expected: controller.ListOrphansResponse{
				Files: []string{"app_id/garbage"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			c := gomock.NewController(t)
			defer c.Finish()

			metadataStorage := mock.NewMockMetadataStorage(c)
			contentStorage := mock.NewMockContentStorage(c)

			metadataStorage.EXPECT().ListFiles(
				gomock.Any(), gomock.Any(),
			).Return(
				[]controller.FileSummary{
					{
						ID:         "b3b4e653-ca59-412c-a165-92d251c3fe86",
						Name:       "file-1.txt",
						IsUploaded: true,
						BucketID:   "default",
					},
					{
						ID:         "7dc0b0d0-b100-4667-89f1-0434942d9c15",
						Name:       "file-two.txt",
						IsUploaded: true,
						BucketID:   "default",
					},
				}, nil,
			)

			contentStorage.EXPECT().ListFiles(gomock.Any()).Return(
				[]string{
					"app_id/b3b4e653-ca59-412c-a165-92d251c3fe86",
					"app_id/7dc0b0d0-b100-4667-89f1-0434942d9c15",
					"app_id/garbage",
				}, nil,
			)

			ctrl := controller.New(
				"http://asd",
				"/v1",
				"asdasd",
				metadataStorage,
				contentStorage,
				nil,
				nil,
				logger,
			)

			router, _ := ctrl.SetupRouter(nil, "/v1", []string{"*"}, false, ginLogger(logger))

			responseRecorder := httptest.NewRecorder()

			req, _ := http.NewRequestWithContext(
				context.Background(),
				"POST",
				"/v1/ops/list-orphans",
				nil,
			)

			router.ServeHTTP(responseRecorder, req)

			assert(t, 200, responseRecorder.Code)

			resp := &controller.ListOrphansResponse{}
			if err := json.Unmarshal(responseRecorder.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			assert(t, &tc.expected, resp)
		})
	}
}
