// gcsreader implements io.ReaderAt.
// it wraps file object in google cloud storage.
package gcsreader

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
)

type GCSObject struct {
	ctx   context.Context
	o     *storage.ObjectHandle
	attrs *storage.ObjectAttrs
}

// Reader implements io.Reader, io.ReaderAt and Size interface
type Reader interface {
	io.Reader
	io.ReaderAt
	Size() int64
}

const maxZeroCount = 10

var (
	ErrNegativeOffset = errors.New("gcs.Reader.ReadAt: negative offset")
)

// Read reads data into p. It returns the number of bytes read into p.
func (g *GCSObject) Read(p []byte) (int, error) {
	r, err := g.o.NewReader(g.ctx)
	if err != nil {
		return 0, errors.Wrap(err, "gcsreader.Reader.Read: storage.ObjectHandle.Read failed")
	}
	defer r.Close()

	n, err := r.Read(p)
	return n, errors.Wrap(err, "gcsreader.Reader.Read: storage.Reader.Read failed")
}

// ReadAt reads len(b) bytes from the storage.ObjectHandle starting at byte offset off.
// It returns the number of bytes read and the error, if any. ReadAt always returns a non-nil error when n < len(b). At end of file, that error is io.EOF.
func (g *GCSObject) ReadAt(b []byte, off int64) (int, error) {
	if off < 0 {
		return 0, ErrNegativeOffset
	}
	if len(b) == 0 {
		return 0, nil
	}

	r, err := g.o.NewRangeReader(g.ctx, off, int64(len(b)))
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer r.Close()

	bufSize := int64(len(b))
	overSize := (off + bufSize) - g.attrs.Size
	if overSize > 0 {
		b = b[:bufSize-overSize]
	}

	n := 0
	errs := errors.New("gcsreader.Reader.ReadAt: storage.Reader.Read failed")
	for i, zc := 0, 0; 0 < len(b) || zc < maxZeroCount; i++ {
		var m int
		m, err = r.Read(b)
		if m == 0 {
			zc++
		} else {
			zc = 0
		}
		if err != nil {
			msg := fmt.Sprintf("[%d] io.Read(b:len(b)=%d) returns m:%d, err:%s", i, len(b), m, err.Error())
			errs = errors.Wrap(errs, msg)
		}
		n = n + m
		b = b[m:]
	}
	if len(b) != 0 {
		return n, errs
	}

	if overSize > 0 {
		return n, io.EOF
	}

	return n, nil
}

// Size returns the size of file in google cloud storage.
func (g *GCSObject) Size() int64 {
	return g.attrs.Size
}

// NewReader returns a new Reader. it wraps storage.ObjectHandle.
func NewReader(ctx context.Context, o *storage.ObjectHandle) (Reader, error) {
	attrs, err := o.Attrs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "gcsreader.NewReader: storage.ObjectHandle.Attrs failed")
	}

	r := &GCSObject{
		ctx:   ctx,
		o:     o,
		attrs: attrs,
	}

	return r, nil
}
