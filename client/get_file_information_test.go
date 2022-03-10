//go:build integration

package client_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/nhost/hasura-storage/client"
	"github.com/nhost/hasura-storage/controller"
	"golang.org/x/net/context"
)

func TestGetFileInformation(t *testing.T) {
	baseURL := "http://localhost:8000/v1/storage"
	cl := client.New(baseURL, os.Getenv("HASURA_AUTH_BEARER"))

	files := []fileHelper{
		{
			name: randomString(),
			path: "testdata/alphabet.txt",
			id:   uuid.NewString(),
		},
		{
			name: randomString(),
			path: "testdata/nhost.jpg",
			id:   uuid.NewString(),
		},
	}
	testFiles, err := uploadFiles(t, cl, files...)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name        string
		id          string
		expected    *client.FileInformationHeader
		expectedErr error
		opts        []client.GetFileInformationOpt
	}{
		{
			name: "get file information, if-match==etag",
			id:   testFiles.ProcessedFiles[0].ID,
			opts: []client.GetFileInformationOpt{client.WithIfMatch(testFiles.ProcessedFiles[0].ETag)},
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 63,
				ContentType:   "text/plain; charset=utf-8",
				Etag:          `"588be441fe7a59460850b0aa3e1c5a65"`,
				LastModified:  "Tue, 18 Jan 2022 13:18:04 UTC",
				StatusCode:    200,
			},
		},
		{
			name: "get file information, if-match!=etag",
			id:   testFiles.ProcessedFiles[0].ID,
			opts: []client.GetFileInformationOpt{client.WithIfMatch("garbage")},
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 63,
				ContentType:   "text/plain; charset=utf-8",
				Etag:          `"588be441fe7a59460850b0aa3e1c5a65"`,
				LastModified:  "Tue, 18 Jan 2022 13:18:04 UTC",
				StatusCode:    412,
			},
		},
		{
			name: "get file information, if-none-match==etag",
			id:   testFiles.ProcessedFiles[0].ID,
			opts: []client.GetFileInformationOpt{client.WithNoneMatch(testFiles.ProcessedFiles[0].ETag)},
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 0,
				ContentType:   "",
				Etag:          `"588be441fe7a59460850b0aa3e1c5a65"`,
				LastModified:  "Tue, 18 Jan 2022 13:18:04 UTC",
				StatusCode:    304,
			},
		},
		{
			name: "get file information, if-none-match!=etag",
			id:   testFiles.ProcessedFiles[0].ID,
			opts: []client.GetFileInformationOpt{client.WithNoneMatch("garbage")},
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 63,
				ContentType:   "text/plain; charset=utf-8",
				Etag:          `"588be441fe7a59460850b0aa3e1c5a65"`,
				LastModified:  "Tue, 18 Jan 2022 13:18:04 UTC",
				StatusCode:    200,
			},
		},
		{
			name: "get file information, if-modified-since!=date",
			id:   testFiles.ProcessedFiles[0].ID,
			opts: []client.GetFileInformationOpt{client.WithIfModifiedSince("Thu, 23 Dec 2025 10:00:00 UTC")},
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 0,
				ContentType:   "",
				Etag:          `"588be441fe7a59460850b0aa3e1c5a65"`,
				LastModified:  "Tue, 18 Jan 2022 13:18:04 UTC",
				StatusCode:    304,
			},
		},
		{
			name: "get file information, if-modified-since==date",
			id:   testFiles.ProcessedFiles[0].ID,
			opts: []client.GetFileInformationOpt{client.WithIfModifiedSince("Thu, 23 Dec 2020 10:00:00 UTC")},
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 63,
				ContentType:   "text/plain; charset=utf-8",
				Etag:          `"588be441fe7a59460850b0aa3e1c5a65"`,
				LastModified:  "Tue, 18 Jan 2022 13:18:04 UTC",
				StatusCode:    200,
			},
		},
		{
			name: "get file information, if-unmodified-since!=date",
			id:   testFiles.ProcessedFiles[0].ID,
			opts: []client.GetFileInformationOpt{client.WithIfUnmodifiedSince("Thu, 23 Dec 2025 10:00:00 UTC")},
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 63,
				ContentType:   "text/plain; charset=utf-8",
				Etag:          `"588be441fe7a59460850b0aa3e1c5a65"`,
				LastModified:  "Tue, 18 Jan 2022 13:18:04 UTC",
				StatusCode:    200,
			},
		},
		{
			name: "get file information, if-unmodified-since==date",
			id:   testFiles.ProcessedFiles[0].ID,
			opts: []client.GetFileInformationOpt{client.WithIfUnmodifiedSince("Thu, 23 Dec 2020 10:00:00 UTC")},
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 63,
				ContentType:   "text/plain; charset=utf-8",
				Etag:          `"588be441fe7a59460850b0aa3e1c5a65"`,
				LastModified:  "Tue, 18 Jan 2022 13:18:04 UTC",
				StatusCode:    412,
			},
		},
		{
			name: "bad id",
			id:   "asdadasdads",
			expectedErr: &client.APIResponseError{
				StatusCode: 400,
				ErrorResponse: &controller.ErrorResponse{
					Message: `Message: invalid input syntax for type uuid: "asdadasdads", Locations: []`,
				},
				Response: nil,
			},
		},
		{
			name: "not found",
			id:   "93aa5806-3050-4810-817a-c917245bb6c1",
			expectedErr: &client.APIResponseError{
				StatusCode: 404,
				ErrorResponse: &controller.ErrorResponse{
					Message: "file not found",
				},
				Response: nil,
			},
		},
		{
			name: "image",
			id:   testFiles.ProcessedFiles[1].ID,
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 33399,
				ContentType:   "image/jpeg",
				Etag:          `"78b676e65ebc31f0bb1f2f0d05098572"`,
				LastModified:  "",
				Error:         "",
				StatusCode:    200,
			},
			expectedErr: nil,
			opts:        []client.GetFileInformationOpt{},
		},
		{
			name: "image/blur",
			id:   testFiles.ProcessedFiles[1].ID,
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 25860,
				ContentType:   "image/jpeg",
				Etag:          `"1c09c3c99b93297f4ea53174d4fb1f5e493faae59e1a106ace855ae316bf7d79"`,
				LastModified:  "",
				Error:         "",
				StatusCode:    200,
			},
			expectedErr: nil,
			opts: []client.GetFileInformationOpt{
				client.WithImageBlur(2),
			},
		},
		{
			name: "image/resized",
			id:   testFiles.ProcessedFiles[1].ID,
			expected: &client.FileInformationHeader{
				CacheControl:  "max-age=3600",
				ContentLength: 10802,
				ContentType:   "image/jpeg",
				Etag:          `"46a94212e38b1bb3fbf769c4bd3c2ff90ccee38ca21ad635281f90c1a316eada"`,
				LastModified:  "",
				Error:         "",
				StatusCode:    200,
			},
			expectedErr: nil,
			opts: []client.GetFileInformationOpt{
				client.WithImageSize(200, 200),
				client.WithImageQuality(90),
			},
		},
		{
			name: "get text file manipulated",
			id:   testFiles.ProcessedFiles[0].ID,
			opts: []client.GetFileInformationOpt{
				client.WithImageSize(600, 200),
				client.WithImageQuality(50),
				client.WithImageBlur(5),
			},
			expectedErr: &client.APIResponseError{
				StatusCode: 400,
				ErrorResponse: &controller.ErrorResponse{
					Message: "image manipulation features are not supported for 'text/plain; charset=utf-8'",
				},
				Response: nil,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			got, err := cl.GetFileInformation(context.Background(), tc.id, tc.opts...)

			if !cmp.Equal(err, tc.expectedErr) {
				t.Errorf(cmp.Diff(err, tc.expectedErr))
			}

			copts := cmp.Options{
				cmpopts.IgnoreFields(client.FileInformationHeader{}, "LastModified"),
			}

			if !cmp.Equal(got, tc.expected, copts...) {
				t.Errorf(cmp.Diff(got, tc.expected, copts...))
			}
		})
	}
}
