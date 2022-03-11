//go:build integration

package client_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/nhost/hasura-storage/client"
	"github.com/nhost/hasura-storage/controller"
	"golang.org/x/net/context"
)

func TestListBucketFiles(t *testing.T) {
	baseURL := "http://localhost:8000/v1/storage"
	cl := client.New(baseURL, os.Getenv("HASURA_AUTH_BEARER"))

	id1 := uuid.NewString()
	id2 := uuid.NewString()
	id3 := uuid.NewString()
	prefix := randomString()

	files := []fileHelper{
		{
			name: randomizedName(prefix, "alphabet.txt"),
			path: "testdata/alphabet.txt",
			id:   id1,
		},
		{
			name: randomizedName(prefix, "asd.txt"),
			path: "testdata/alphabet.txt",
			id:   id2,
		},
		{
			name: randomizedName(prefix, "txt"),
			path: "testdata/alphabet.txt",
			id:   id3,
		},
	}

	_, err := uploadFiles(t, cl, files...)
	if err != nil {
		panic(err)
	}

	cases := []struct {
		name        string
		id          string
		filter      string
		cl          *client.Client
		expected    *controller.ListBucketFilesResponse
		expectedErr error
	}{
		{
			name:   "filter",
			id:     "default",
			filter: fmt.Sprintf(`^%s`, prefix),
			cl:     cl,
			expected: &controller.ListBucketFilesResponse{
				Files: []controller.FileMetadata{
					{
						ID:         "e650c224-05ea-4144-9fb5-2f086623c9c4",
						Name:       prefix + "-alphabet.txt",
						Size:       63,
						BucketID:   "default",
						ETag:       `"588be441fe7a59460850b0aa3e1c5a65"`,
						CreatedAt:  "2022-03-10T12:35:44.522263+00:00",
						UpdatedAt:  "2022-03-10T12:35:44.541153+00:00",
						IsUploaded: true,
						MimeType:   "text/plain; charset=utf-8",
					},
					{
						ID:         "e7ed223d-0898-4ff7-8897-7103e314d38d",
						Name:       prefix + "-asd.txt",
						Size:       63,
						BucketID:   "default",
						ETag:       `"588be441fe7a59460850b0aa3e1c5a65"`,
						CreatedAt:  "2022-03-10T12:35:44.548067+00:00",
						UpdatedAt:  "2022-03-10T12:35:44.557396+00:00",
						IsUploaded: true,
						MimeType:   "text/plain; charset=utf-8",
					},
					{
						ID:         "44ea277c-b32f-4656-9307-3592fde71606",
						Name:       prefix + "-txt",
						Size:       63,
						BucketID:   "default",
						ETag:       `"588be441fe7a59460850b0aa3e1c5a65"`,
						CreatedAt:  "2022-03-10T12:35:44.561193+00:00",
						UpdatedAt:  "2022-03-10T12:35:44.567806+00:00",
						IsUploaded: true,
						MimeType:   "text/plain; charset=utf-8",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "not found",
			id:   "askjnmbsd",
			cl:   cl,
			expected: &controller.ListBucketFilesResponse{
				Error: &controller.ErrorResponse{
					Message: "bucket not found",
				},
			},
			expectedErr: nil,
		},
		{
			name: "not authorized",
			id:   "default",
			cl:   client.New(baseURL, ""),
			expected: &controller.ListBucketFilesResponse{
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

			got, err := tc.cl.ListBucketFiles(context.Background(), tc.id, tc.filter)
			if !cmp.Equal(err, tc.expectedErr) {
				t.Errorf(cmp.Diff(err, tc.expectedErr))
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(controller.FileMetadata{}, "ID", "CreatedAt", "UpdatedAt"),
			}

			if !cmp.Equal(got, tc.expected, opts...) {
				t.Errorf(cmp.Diff(got, tc.expected, opts...))
			}
		})
	}
}
