package image

import (
	"bytes"
	"fmt"
	"io"

	"github.com/davidbyttow/govips/v2/vips"
)

const (
	poolSize = 4
)

var buffs = newBufferPool()

type bufferPool struct {
	ch chan struct{}
}

func newBufferPool() *bufferPool {
	ch := make(chan struct{}, poolSize)

	for i := 0; i < poolSize; i++ {
		ch <- struct{}{}
	}
	return &bufferPool{
		ch: ch,
	}
}

func (p *bufferPool) get() *bytes.Buffer {
	<-p.ch
	return bytes.NewBuffer(make([]byte, 0, 5*1024*1024))
}

func (p *bufferPool) put(buf *bytes.Buffer) {
	p.ch <- struct{}{}
}

type Options func(orig *vips.ImageRef, params *vips.ExportParams)

func WithNewSize(x, y int) Options { // nolint: varnamelen
	return func(orig *vips.ImageRef, _ *vips.ExportParams) {
		if x == 0 {
			x = orig.Width() * y / orig.Height()
		}
		if y == 0 {
			y = orig.Height() * x / orig.Width()
		}
		if err := orig.Thumbnail(x, y, vips.InterestingCentre); err != nil {
			panic(err)
		}
	}
}

func WithBlur(sigma float64) Options {
	return func(orig *vips.ImageRef, _ *vips.ExportParams) {
		if err := orig.GaussianBlur(sigma); err != nil {
			panic(err)
		}
	}
}

func WithQuality(q int) Options {
	return func(_ *vips.ImageRef, params *vips.ExportParams) {
		params.Quality = q
	}
}

func Manipulate(orig io.Reader, modified io.Writer, opts ...Options) error {
	buf := buffs.get()
	defer buffs.put(buf)

	_, err := io.Copy(buf, orig)
	if err != nil {
		panic(err)
	}

	image1, err := vips.NewImageFromBuffer(buf.Bytes())
	if err != nil {
		return fmt.Errorf("problem creating image from reader: %w", err)
	}
	defer image1.Close()

	params := &vips.ExportParams{}

	for _, o := range opts {
		o(image1, params)
	}

	b, _, err := image1.Export(params)
	if err != nil {
		return fmt.Errorf("problem exporting image: %w", err)
	}

	if _, err := modified.Write(b); err != nil {
		return fmt.Errorf("problem writing image: %w", err)
	}

	return nil
}
