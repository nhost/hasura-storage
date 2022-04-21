//go:build integration
// +build integration

package metadata_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/nhost/hasura-storage/controller"
	"github.com/nhost/hasura-storage/metadata"
)

const (
	hasuraURL = "http://localhost:8080/v1/graphql"
)

func getAuthHeader() http.Header {
	headers := http.Header{}
	bearer := os.Getenv("HASURA_AUTH_BEARER")
	headers.Add("Authorization", "Bearer "+bearer)

	return headers
}

func TestGetBucketByID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                   string
		bucketID               string
		headers                http.Header
		expectedStatusCode     int
		expectedPublicResponse *controller.ErrorResponse
		expected               controller.BucketMetadata
	}{
		{
			name:                   "success",
			bucketID:               "default",
			headers:                getAuthHeader(),
			expectedStatusCode:     0,
			expectedPublicResponse: &controller.ErrorResponse{},
			expected: controller.BucketMetadata{
				ID:                   "default",
				MinUploadFile:        1,
				MaxUploadFile:        50000000,
				PresignedURLsEnabled: true,
				DownloadExpiration:   30,
				CreatedAt:            "",
				UpdatedAt:            "",
				CacheControl:         "max-age=3600",
			},
		},
		{
			name:               "not found",
			bucketID:           "asdsad",
			headers:            getAuthHeader(),
			expectedStatusCode: 404,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "bucket not found",
			},
			expected: controller.BucketMetadata{},
		},
		{
			name:               "not authorized",
			bucketID:           "asdsad",
			expectedStatusCode: 403,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "you are not authorized",
			},
			expected: controller.BucketMetadata{},
		},
	}

	hasura := metadata.NewHasura(hasuraURL, metadata.ForWardHeadersAuthorizer)

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			bucket, err := hasura.GetBucketByID(context.Background(), tc.bucketID, tc.headers)

			if tc.expectedStatusCode != err.StatusCode() {
				t.Errorf("wrong status code, expected %d, got %d", tc.expectedStatusCode, err.StatusCode())
			}

			if err != nil {
				if !cmp.Equal(err.PublicResponse(), tc.expectedPublicResponse) {
					t.Error(cmp.Diff(err.PublicResponse(), tc.expectedPublicResponse))
				}
			} else {
				opts := cmp.Options{
					cmpopts.IgnoreFields(controller.BucketMetadata{}, "CreatedAt", "UpdatedAt"),
				}
				if !cmp.Equal(bucket, tc.expected, opts...) {
					t.Error(cmp.Diff(bucket, tc.expected, opts...))
				}
			}
		})
	}
}

func TestInitializeFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                   string
		fileID                 string
		headers                http.Header
		expectedStatusCode     int
		expectedPublicResponse *controller.ErrorResponse
	}{
		{
			name:                   "success",
			fileID:                 uuid.New().String(),
			headers:                getAuthHeader(),
			expectedStatusCode:     0,
			expectedPublicResponse: &controller.ErrorResponse{},
		},
		{
			name:               "wrong format",
			fileID:             "asdsad",
			headers:            getAuthHeader(),
			expectedStatusCode: 400,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "Message: invalid input syntax for type uuid: \"asdsad\", Locations: []",
			},
		},
		{
			name:               "not authorized",
			fileID:             "asdsad",
			expectedStatusCode: 403,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "you are not authorized",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			hasura := metadata.NewHasura(hasuraURL, metadata.ForWardHeadersAuthorizer)

			err := hasura.InitializeFile(context.Background(), tc.fileID, tc.headers)

			if tc.expectedStatusCode != err.StatusCode() {
				t.Errorf("wrong status code, expected %d, got %d", tc.expectedStatusCode, err.StatusCode())
			}
			if err != nil {
				if !cmp.Equal(err.PublicResponse(), tc.expectedPublicResponse) {
					t.Error(cmp.Diff(err.PublicResponse(), tc.expectedPublicResponse))
				}
			}
		})
	}
}

func TestPopulateMetadata(t *testing.T) {
	t.Parallel()

	hasura := metadata.NewHasura(hasuraURL, metadata.ForWardHeadersAuthorizer)

	fileID := uuid.New().String()
	if err := hasura.InitializeFile(context.Background(), fileID, getAuthHeader()); err != nil {
		panic(err)
	}

	cases := []struct {
		name                   string
		fileID                 string
		headers                http.Header
		expectedStatusCode     int
		expectedPublicResponse *controller.ErrorResponse
		expected               controller.FileMetadata
	}{
		{
			name:                   "success",
			fileID:                 fileID,
			headers:                getAuthHeader(),
			expectedStatusCode:     0,
			expectedPublicResponse: &controller.ErrorResponse{},
			expected: controller.FileMetadata{
				ID:               fileID,
				Name:             "name",
				Size:             123,
				BucketID:         "default",
				ETag:             "asdasd",
				CreatedAt:        "",
				UpdatedAt:        "",
				IsUploaded:       true,
				MimeType:         "text",
				UploadedByUserID: "",
			},
		},
		{
			name:               "wrong format",
			fileID:             "asdasdasd",
			headers:            getAuthHeader(),
			expectedStatusCode: 400,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: `Message: invalid input syntax for type uuid: "asdasdasd", Locations: []`,
			},
			expected: controller.FileMetadata{},
		},
		{
			name:               "not found",
			fileID:             uuid.New().String(),
			headers:            getAuthHeader(),
			expectedStatusCode: 404,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "file not found",
			},
			expected: controller.FileMetadata{},
		},
		{
			name:               "not authorized",
			expectedStatusCode: 403,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "you are not authorized",
			},
			expected: controller.FileMetadata{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			got, err := hasura.PopulateMetadata(context.Background(), tc.fileID, "name", 123, "default", "asdasd", true, "text", tc.headers)

			if tc.expectedStatusCode != err.StatusCode() {
				t.Errorf("wrong status code, expected %d, got %d", tc.expectedStatusCode, err.StatusCode())
			}
			if err != nil {
				if !cmp.Equal(err.PublicResponse(), tc.expectedPublicResponse) {
					t.Error(cmp.Diff(err.PublicResponse(), tc.expectedPublicResponse))
				}
			} else {
				opts := cmp.Options{
					cmpopts.IgnoreFields(controller.FileMetadata{}, "CreatedAt", "UpdatedAt"),
				}
				if !cmp.Equal(got, tc.expected, opts...) {
					t.Error(cmp.Diff(got, tc.expected, opts...))
				}
			}
		})
	}
}

func TestGetFileByID(t *testing.T) {
	t.Parallel()

	hasura := metadata.NewHasura(hasuraURL, metadata.ForWardHeadersAuthorizer)

	fileID := uuid.New().String()
	if err := hasura.InitializeFile(context.Background(), fileID, getAuthHeader()); err != nil {
		panic(err)
	}

	if _, err := hasura.PopulateMetadata(context.Background(), fileID, "name", 123, "default", "asdasd", true, "text", getAuthHeader()); err != nil {
		panic(err)
	}

	cases := []struct {
		name                   string
		fileID                 string
		headers                http.Header
		expectedStatusCode     int
		expectedPublicResponse *controller.ErrorResponse
		expected               controller.FileMetadata
	}{
		{
			name:                   "success",
			fileID:                 fileID,
			headers:                getAuthHeader(),
			expectedStatusCode:     0,
			expectedPublicResponse: &controller.ErrorResponse{},
			expected: controller.FileMetadata{
				ID:               fileID,
				Name:             "name",
				Size:             123,
				BucketID:         "default",
				ETag:             "asdasd",
				CreatedAt:        "",
				UpdatedAt:        "",
				IsUploaded:       true,
				MimeType:         "text",
				UploadedByUserID: "",
			},
		},
		{
			name:               "wrong format",
			fileID:             "asdasdasd",
			headers:            getAuthHeader(),
			expectedStatusCode: 400,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: `Message: invalid input syntax for type uuid: "asdasdasd", Locations: []`,
			},
			expected: controller.FileMetadata{},
		},
		{
			name:               "not found",
			fileID:             uuid.New().String(),
			headers:            getAuthHeader(),
			expectedStatusCode: 404,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "file not found",
			},
			expected: controller.FileMetadata{},
		},
		{
			name:               "not authorized",
			expectedStatusCode: 403,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "you are not authorized",
			},
			expected: controller.FileMetadata{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			got, err := hasura.GetFileByID(context.Background(), tc.fileID, tc.headers)

			if tc.expectedStatusCode != err.StatusCode() {
				t.Errorf("wrong status code, expected %d, got %d", tc.expectedStatusCode, err.StatusCode())
			}
			if err != nil {
				if !cmp.Equal(err.PublicResponse(), tc.expectedPublicResponse) {
					t.Error(cmp.Diff(err.PublicResponse(), tc.expectedPublicResponse))
				}
			} else {
				opts := cmp.Options{
					cmpopts.IgnoreFields(controller.FileMetadata{}, "CreatedAt", "UpdatedAt"),
					cmpopts.IgnoreFields(controller.BucketMetadata{}, "CreatedAt", "UpdatedAt"),
				}
				if !cmp.Equal(got, tc.expected, opts...) {
					t.Error(cmp.Diff(got, tc.expected, opts...))
				}
			}
		})
	}
}

func TestSetIsUploaded(t *testing.T) {
	t.Parallel()

	hasura := metadata.NewHasura(hasuraURL, metadata.ForWardHeadersAuthorizer)

	fileID := uuid.New().String()
	if err := hasura.InitializeFile(context.Background(), fileID, getAuthHeader()); err != nil {
		panic(err)
	}

	cases := []struct {
		name                   string
		fileID                 string
		headers                http.Header
		expectedStatusCode     int
		expectedPublicResponse *controller.ErrorResponse
	}{
		{
			name:                   "success",
			fileID:                 fileID,
			headers:                getAuthHeader(),
			expectedStatusCode:     0,
			expectedPublicResponse: &controller.ErrorResponse{},
		},
		{
			name:               "file not found",
			fileID:             "aaaaaaaa-1111-bbbb-2222-cccccccccccc",
			headers:            getAuthHeader(),
			expectedStatusCode: http.StatusNotFound,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "file not found",
			},
		},
		{
			name:               "not authorized",
			fileID:             "aaaaaaaa-1111-bbbb-2222-cccccccccccc",
			headers:            map[string][]string{},
			expectedStatusCode: http.StatusForbidden,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "you are not authorized",
			},
		},
		{
			name:               "wrong id",
			fileID:             "",
			headers:            getAuthHeader(),
			expectedStatusCode: http.StatusBadRequest,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "Message: invalid input syntax for type uuid: \"\", Locations: []",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			err := hasura.SetIsUploaded(context.Background(), tc.fileID, true, tc.headers)
			if tc.expectedStatusCode != err.StatusCode() {
				t.Errorf("wrong status code, expected %d, got %d", tc.expectedStatusCode, err.StatusCode())
			}

			if err != nil {
				if !cmp.Equal(err.PublicResponse(), tc.expectedPublicResponse) {
					t.Error(cmp.Diff(err.PublicResponse(), tc.expectedPublicResponse))
				}
			}
		})
	}
}

func TestDeleteFileByID(t *testing.T) {
	t.Parallel()

	hasura := metadata.NewHasura(hasuraURL, metadata.ForWardHeadersAuthorizer)

	fileID := uuid.New().String()
	if err := hasura.InitializeFile(context.Background(), fileID, getAuthHeader()); err != nil {
		panic(err)
	}

	if _, err := hasura.PopulateMetadata(context.Background(), fileID, "name", 123, "default", "asdasd", true, "text", getAuthHeader()); err != nil {
		panic(err)
	}

	cases := []struct {
		name                   string
		fileID                 string
		headers                http.Header
		expected               controller.FileMetadataWithBucket
		expectedStatusCode     int
		expectedPublicResponse *controller.ErrorResponse
	}{
		{
			name:    "success",
			fileID:  fileID,
			headers: getAuthHeader(),
			expected: controller.FileMetadataWithBucket{
				FileMetadata: controller.FileMetadata{
					ID:               fileID,
					Name:             "name",
					Size:             123,
					BucketID:         "default",
					ETag:             "asdasd",
					CreatedAt:        "",
					UpdatedAt:        "",
					IsUploaded:       true,
					MimeType:         "text",
					UploadedByUserID: "",
				},
				Bucket: controller.BucketMetadata{
					ID:                   "default",
					MinUploadFile:        1,
					MaxUploadFile:        50000000,
					PresignedURLsEnabled: true,
					DownloadExpiration:   30,
					CreatedAt:            "2022-01-05T19:02:58.387709+00:00",
					UpdatedAt:            "2022-01-05T19:02:58.387709+00:00",
					CacheControl:         "max-age=3600",
				},
			},
			expectedStatusCode:     0,
			expectedPublicResponse: &controller.ErrorResponse{},
		},
		{
			name:               "file not found",
			fileID:             "aaaaaaaa-1111-bbbb-2222-cccccccccccc",
			headers:            getAuthHeader(),
			expectedStatusCode: http.StatusNotFound,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "file not found",
			},
		},
		{
			name:               "not authorized",
			fileID:             "aaaaaaaa-1111-bbbb-2222-cccccccccccc",
			headers:            map[string][]string{},
			expectedStatusCode: http.StatusForbidden,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "you are not authorized",
			},
		},
		{
			name:               "wrong id",
			fileID:             "",
			headers:            getAuthHeader(),
			expectedStatusCode: http.StatusBadRequest,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "Message: invalid input syntax for type uuid: \"\", Locations: []",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			got, err := hasura.DeleteFileByID(context.Background(), tc.fileID, tc.headers)
			if tc.expectedStatusCode != err.StatusCode() {
				t.Errorf("wrong status code, expected %d, got %d", tc.expectedStatusCode, err.StatusCode())
			}
			if err != nil {
				if !cmp.Equal(err.PublicResponse(), tc.expectedPublicResponse) {
					t.Error(cmp.Diff(err.PublicResponse(), tc.expectedPublicResponse))
				}
			} else {
				opts := cmp.Options{
					cmpopts.IgnoreFields(controller.FileMetadata{}, "CreatedAt", "UpdatedAt"),
					cmpopts.IgnoreFields(controller.BucketMetadata{}, "CreatedAt", "UpdatedAt"),
				}
				if !cmp.Equal(got, tc.expected, opts...) {
					t.Error(cmp.Diff(got, tc.expected, opts...))
				}
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	hasura := metadata.NewHasura(hasuraURL, metadata.ForWardHeadersAuthorizer)

	fileID1 := uuid.New().String()
	if err := hasura.InitializeFile(context.Background(), fileID1, getAuthHeader()); err != nil {
		panic(err)
	}

	if _, err := hasura.PopulateMetadata(context.Background(), fileID1, "name", 123, "default", "asdasd", true, "text", getAuthHeader()); err != nil {
		panic(err)
	}

	fileID2 := uuid.New().String()
	if err := hasura.InitializeFile(context.Background(), fileID2, getAuthHeader()); err != nil {
		panic(err)
	}

	if _, err := hasura.PopulateMetadata(context.Background(), fileID2, "asdads", 123, "default", "asdasd", true, "text", getAuthHeader()); err != nil {
		panic(err)
	}

	cases := []struct {
		name                   string
		headers                http.Header
		expectedStatusCode     int
		expectedPublicResponse *controller.ErrorResponse
	}{
		{
			name:                   "success",
			headers:                getAuthHeader(),
			expectedStatusCode:     0,
			expectedPublicResponse: &controller.ErrorResponse{},
		},
		{
			name:               "unauthorized",
			headers:            map[string][]string{},
			expectedStatusCode: 403,
			expectedPublicResponse: &controller.ErrorResponse{
				Message: "you are not authorized",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			got, err := hasura.ListFiles(context.Background(), tc.headers)
			if tc.expectedStatusCode != err.StatusCode() {
				t.Errorf("wrong status code, expected %d, got %d", tc.expectedStatusCode, err.StatusCode())
			}

			if err != nil {
				if !cmp.Equal(err.PublicResponse(), tc.expectedPublicResponse) {
					t.Error(cmp.Diff(err.PublicResponse(), tc.expectedPublicResponse))
				}
			} else {
				if len(got) == 0 {
					t.Error("we got an empty list")
				}

				found1 := false
				found2 := false
				for _, f := range got {
					if f.ID == fileID1 {
						found1 = true
					} else if f.ID == fileID2 {
						found2 = true
					}
				}
				if !found1 || !found2 {
					t.Error("couldn't find some files in the list")
				}
			}
		})
	}
}
