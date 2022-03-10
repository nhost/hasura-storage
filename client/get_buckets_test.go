//go:build integration

package client_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nhost/hasura-storage/client"
	"github.com/nhost/hasura-storage/controller"
	"golang.org/x/net/context"
)

func TestGetBuckets(t *testing.T) {
	baseURL := "http://localhost:8000/v1/storage"
	cl := client.New(baseURL, os.Getenv("HASURA_AUTH_BEARER"))

	cases := []struct {
		name        string
		cl          *client.Client
		expected    *controller.GetBucketsResponse
		expectedErr error
	}{
		{
			name: "success",
			cl:   cl,
			expected: &controller.GetBucketsResponse{
				Buckets: []controller.BucketMetadata{
					{
						ID:                   "default",
						MinUploadFile:        1,
						MaxUploadFile:        50000000,
						PresignedURLsEnabled: true,
						DownloadExpiration:   30,
						CreatedAt:            "2022-03-10T10:07:48.739504+00:00",
						UpdatedAt:            "2022-03-10T10:07:48.739504+00:00",
						CacheControl:         "max-age=3600",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "not authorized",
			cl:   client.New(baseURL, ""),
			expected: &controller.GetBucketsResponse{
				Error: &controller.ErrorResponse{
					Message: "Message: Malformed Authorization header, Locations: []",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			got, err := tc.cl.GetBuckets(context.Background())
			if !cmp.Equal(err, tc.expectedErr) {
				t.Errorf(cmp.Diff(err, tc.expectedErr))
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(controller.BucketMetadata{}, "CreatedAt", "UpdatedAt"),
			}

			if !cmp.Equal(got, tc.expected, opts...) {
				t.Errorf(cmp.Diff(got, tc.expected, opts...))
			}
		})
	}
}
