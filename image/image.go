package image

// #cgo pkg-config: vips
// #include <vips/vips.h>
// #include "image.h"
import "C"

import (
	"bytes"
	"fmt"
	"io"
	"runtime/debug"
	"unsafe"
)

const (
	maxWorkers = 3
	buffSize   = 5 << 20
)

type ImageType int

const (
	ImageTypeJPEG ImageType = C.JPEG
	ImageTypePNG  ImageType = C.PNG
	ImageTypeWEBP ImageType = C.WEBP
)

type Options struct {
	Height  int
	Width   int
	Blur    float64
	Quality int
	Format  ImageType
}

func (o Options) IsEmpty() bool {
	return o.Height == 0 && o.Width == 0 && o.Blur == 0 && o.Quality == 0
}

type Transformer struct {
	workers chan struct{}
}

func NewTransformer() *Transformer {
	name := C.CString("hasuraStorage")
	defer C.free(unsafe.Pointer(name))

	err := C.vips_init(name)
	if err != 0 {
		panic(fmt.Sprintf("vips error, code=%v", err))
	}

	C.vips_concurrency_set(C.int(1))
	C.vips_cache_set_max_files(C.int(0))
	C.vips_cache_set_max_mem(C.size_t(0))
	C.vips_cache_set_max(C.int(0))

	workers := make(chan struct{}, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers <- struct{}{}
	}
	return &Transformer{workers: workers}
}

func (t *Transformer) Shutdown() {
	C.vips_shutdown()
}

func (t *Transformer) Run(orig io.Reader, modified io.Writer, opts Options) error {
	// this is to avoid processing too many images at the same time in order to save memory
	<-t.workers
	defer func() { t.workers <- struct{}{} }()

	buf := bytes.NewBuffer(make([]byte, 0, buffSize))

	_, err := io.Copy(buf, orig)
	if err != nil {
		panic(err)
	}

	b, err := Manipulate(buf.Bytes(), opts)
	if err != nil {
		return fmt.Errorf("problem manipulating image: %w", err)
	}

	if _, err := modified.Write(b); err != nil {
		return fmt.Errorf("problem writing image: %w", err)
	}

	return nil
}

func Manipulate(buf []byte, opts Options) ([]byte, error) {
	var result C.Result

	err := C.manipulate(
		unsafe.Pointer(&buf[0]),
		C.size_t(len(buf)),
		&result,
		C.Options{
			width:  C.int(opts.Width),
			height: C.int(opts.Height),
			crop:   C.VIPS_INTERESTING_CENTRE,
			size:   C.VIPS_SIZE_BOTH, // nolint: gocritic
			blur:   C.double(opts.Blur),
			format: C.ImageType(opts.Format),
		},
	)
	if err != 0 {
		s := C.GoString(C.vips_error_buffer())
		C.vips_error_clear()

		return nil, fmt.Errorf("%v\nStack:\n%s", s, debug.Stack()) // nolint: goerr113
	}

	return C.GoBytes(result.buf, C.int(result.len)), nil
}
