package image_test

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nhost/hasura-storage/image"
)

func TestManipulate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		filename string
		sum      string
		size     uint64
		options  image.Options
	}{
		{
			name:     "jpg",
			filename: "testdata/nhost.jpg",
			sum:      "85f4398b93c70cf3c46589cbd4134e4ba65a85736d23e4bbed4e3f34b343028b",
			size:     33399,
			options: image.Options{
				Height:  100,
				Width:   300,
				Blur:    2,
				Quality: 50,
				Format:  image.ImageTypeJPEG,
			},
		},
		{
			name:     "jpg",
			filename: "testdata/nhost.jpg",
			sum:      "c7bad0080993f113e4b1cab8b9e90f9ce4bb5add021ddc0f4d4d915d503cd794",
			size:     33399,
			options:  image.Options{Width: 300, Height: 100, Blur: 2, Format: image.ImageTypeJPEG},
		},
		{
			name:     "png",
			filename: "testdata/nhost.png",
			sum:      "78cf83b463c94ecec430ca424d80650d07be5424929e67cfe978c0c753065745",
			size:     68307,
			options:  image.Options{Width: 300, Height: 100, Blur: 2, Format: image.ImageTypePNG},
		},
		{
			name:     "webp",
			filename: "testdata/nhost.webp",
			sum:      "be74a4ebb7e32fefdf7a901b2f15db0679b13994ce4cc3bfaf2245ccf11dc178",
			size:     17784,
			options:  image.Options{Width: 300, Height: 100, Blur: 2, Format: image.ImageTypeWEBP},
		},
		{
			name:     "jpg only blur",
			filename: "testdata/nhost.jpg",
			sum:      "e7e6b6fb5d949df19d4aa944b3c0d4200c4898616ed4452e064e74c75f601c9b",
			size:     33399,
			options:  image.Options{Blur: 2, Format: image.ImageTypeJPEG},
		},
	}

	transformer := image.NewTransformer()

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			orig, err := os.Open(tc.filename)
			if err != nil {
				t.Fatal(err)
			}
			defer orig.Close()

			hasher := sha256.New()
			// f, _ := os.OpenFile("/tmp/nhost-test."+tc.name, os.O_WRONLY|os.O_CREATE, 0o644)
			if err := transformer.Run(orig, tc.size, hasher, tc.options); err != nil {
				t.Fatal(err)
			}

			got := fmt.Sprintf("%x", hasher.Sum(nil))
			if !cmp.Equal(got, tc.sum) {
				t.Error(cmp.Diff(got, tc.sum))
			}
		})
	}
}

func BenchmarkManipulate(b *testing.B) {
	transformer := image.NewTransformer()
	orig, err := os.Open("testdata/nhost.jpg")
	if err != nil {
		b.Fatal(err)
	}
	defer orig.Close()
	for i := 0; i < 100; i++ {
		_, _ = orig.Seek(0, 0)

		if err := transformer.Run(
			orig,
			33399,
			io.Discard,
			image.Options{Width: 300, Height: 100, Blur: 1.5, Format: image.ImageTypeJPEG},
		); err != nil {
			b.Fatal(err)
		}
	}
}
