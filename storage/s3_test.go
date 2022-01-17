// +build integration

package storage_test

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/google/go-cmp/cmp"
	"github.com/nhost/hasura-storage/storage"
	"github.com/sirupsen/logrus"
)

func getS3() *storage.S3 {
	config := &aws.Config{ // nolint: exhaustivestruct
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("TEST_S3_ACCESS_KEY"),
			os.Getenv("TEST_S3_SECRET_KEY"),
			"",
		),
		Endpoint:         aws.String("http://localhost:9000"),
		Region:           aws.String("eu-central-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}

	logger := logrus.New()

	st, err := storage.NewS3(config, "default", "a-root-folder", logger)
	if err != nil {
		panic(err)
	}
	return st
}

func TestDeleteFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		filepath string
	}{
		{
			name:     "success",
			filepath: "/default/asd",
		},
		{
			name:     "file not found",
			filepath: "/default/qwenmzxcxzcsadsad",
		},
	}
	s3 := getS3()

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc
			err := s3.DeleteFile(tc.filepath)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	cases := []struct {
		name     string
		expected []string
	}{
		{
			name: "success",
			expected: []string{
				"default/asd",
			},
		},
	}
	s3 := getS3()

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			got, err := s3.ListFiles()
			if err != nil {
				t.Error(err)
			}

			if !cmp.Equal(got, tc.expected) {
				t.Error(cmp.Diff(got, tc.expected))
			}
		})
	}
}
