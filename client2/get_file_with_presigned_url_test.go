package client2_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/nhost/hasura-storage/client2"
)

func TestGetPresignedURLContents(t *testing.T) {
	t.Parallel()

	cl, err := client2.NewClientWithResponses(testBaseURL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	id1 := uuid.NewString()
	id2 := uuid.NewString()

	uploadInitialFile(t, cl, id1, id2)

	resp, err := cl.GetFilePresignedURLWithResponse(
		t.Context(), id1, WithAccessToken(accessTokenValidUser),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	presignedURL, err := url.Parse(resp.JSON200.Url)
	if err != nil {
		t.Fatalf("failed to parse presigned URL: %v", err)
	}

	cases := []struct {
		name               string
		id                 string
		requestParams      func(req *client2.GetPresignedURLContentsParams) *client2.GetPresignedURLContentsParams
		interceptor        func(ctx context.Context, req *http.Request) error
		expectedStatusCode int
		expectedBody       string
		expectedErr        *client2.ErrorResponse
		expectedHeaders    http.Header
	}{
		{
			name:        "simple get",
			id:          id1,
			interceptor: WithAccessToken(accessTokenValidUser),
			requestParams: func(req *client2.GetPresignedURLContentsParams) *client2.GetPresignedURLContentsParams {
				return req
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "Hello, World!",
			expectedHeaders: http.Header{
				"Accept-Ranges":       []string{"bytes"},
				"Cache-Control":       []string{"max-age=29"},
				"Content-Disposition": []string{`inline; filename="testfile.txt"`},
				"Content-Length":      []string{"13"},
				"Content-Type":        []string{"text/plain; charset=utf-8"},
				"Date":                []string{"Mon, 21 Jul 2025 13:24:53 GMT"},
				"Etag":                []string{`"65a8e27d8879283831b664bd8b7f0ad4"`},
				"Last-Modified":       []string{"2025-07-21 13:24:53.586273 +0000 +0000"},
				"Surrogate-Control":   []string{"max-age=29"},
				"Surrogate-Key":       []string{"d505075a-ee28-4a02-b27a-5973fd2ea35f"},
			},
		},
		{
			name:        "IfMatch matches",
			id:          id1,
			interceptor: WithAccessToken(accessTokenValidUser),
			requestParams: func(req *client2.GetPresignedURLContentsParams) *client2.GetPresignedURLContentsParams {
				req.IfMatch = ptr(`"65a8e27d8879283831b664bd8b7f0ad4"`)
				return req
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "Hello, World!",
			expectedHeaders: http.Header{
				"Accept-Ranges":       []string{"bytes"},
				"Cache-Control":       []string{"max-age=29"},
				"Content-Disposition": []string{`inline; filename="testfile.txt"`},
				"Content-Length":      []string{"13"},
				"Content-Type":        []string{"text/plain; charset=utf-8"},
				"Date":                []string{"Mon, 21 Jul 2025 13:24:53 GMT"},
				"Etag":                []string{`"65a8e27d8879283831b664bd8b7f0ad4"`},
				"Last-Modified":       []string{"2025-07-21 13:24:53.586273 +0000 +0000"},
				"Surrogate-Control":   []string{"max-age=29"},
				"Surrogate-Key":       []string{"d505075a-ee28-4a02-b27a-5973fd2ea35f"},
			},
		},
		{
			name: "IfMatch does not match",
			id:   id1,
			requestParams: func(req *client2.GetPresignedURLContentsParams) *client2.GetPresignedURLContentsParams {
				req.IfMatch = ptr(`"85a8e27d8879283831b664bd8b7f0ad4"`)
				return req
			},
			interceptor:        WithAccessToken(accessTokenValidUser),
			expectedStatusCode: http.StatusPreconditionFailed,
			expectedBody:       "",
			expectedHeaders: http.Header{
				"Cache-Control":     []string{"max-age=29"},
				"Content-Length":    []string{"0"},
				"Date":              []string{"Mon, 21 Jul 2025 13:24:53 GMT"},
				"Etag":              []string{`"65a8e27d8879283831b664bd8b7f0ad4"`},
				"Last-Modified":     []string{"2025-07-21 13:24:53.586273 +0000 +0000"},
				"Surrogate-Control": []string{"max-age=29"},
				"Surrogate-Key":     []string{"d505075a-ee28-4a02-b27a-5973fd2ea35f"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var interceptor []client2.RequestEditorFn
			if tc.interceptor != nil {
				interceptor = []client2.RequestEditorFn{
					tc.interceptor,
				}
			}

			resp, err := cl.GetPresignedURLContentsWithResponse(
				t.Context(),
				tc.id,
				tc.requestParams(
					&client2.GetPresignedURLContentsParams{
						XAmzAlgorithm:     presignedURL.Query().Get("X-Amz-Algorithm"),
						XAmzChecksumMode:  presignedURL.Query().Get("X-Amz-Checksum-Mode"),
						XAmzCredential:    presignedURL.Query().Get("X-Amz-Credential"),
						XAmzDate:          presignedURL.Query().Get("X-Amz-Date"),
						XAmzExpires:       presignedURL.Query().Get("X-Amz-Expires"),
						XId:               presignedURL.Query().Get("x-id"),
						XAmzSignature:     presignedURL.Query().Get("X-Amz-Signature"),
						XAmzSignedHeaders: presignedURL.Query().Get("X-Amz-SignedHeaders"),
					},
				),
				interceptor...,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.StatusCode() != tc.expectedStatusCode {
				t.Errorf(
					"expected status code %d, got %d", tc.expectedStatusCode, resp.StatusCode(),
				)
			}

			if tc.expectedBody != "ignoreme" {
				if diff := cmp.Diff(string(resp.Body), tc.expectedBody); diff != "" {
					t.Errorf("unexpected response body: %s", diff)
				}
			}

			if diff := cmp.Diff(
				resp.HTTPResponse.Header,
				tc.expectedHeaders,
				cmp.Options{
					cmpopts.IgnoreMapEntries(func(key string, _ []string) bool {
						return key == "Date" || key == "Surrogate-Key" || key == "Last-Modified" ||
							strings.HasPrefix(key, "X-B3-")
					}),
				},
			); diff != "" {
				t.Errorf("unexpected response headers: %s", diff)
			}
		})
	}
}
