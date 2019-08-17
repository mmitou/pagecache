package pagecache

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

var (
	PageNotFound = errors.New("page not found")
	InvalidArg   = errors.New("invalid argument")
)

type Page struct {
	index int64
	b     []byte
}

type PageTable struct {
	queue        []*Page
	pageSize     int64
	maxQueueSize int64 // must be greater than 1
}

type PageCache struct {
	ra        io.ReaderAt
	pageTable *PageTable
}

func (pt *PageTable) get(i int64) (*Page, error) {
	if i < 0 {
		return nil, InvalidArg
	}

	q := pt.queue
	qi := -1
	for j := len(q) - 1; j >= 0; j-- {
		if q[j].index == i {
			qi = j
		}
	}

	if qi == -1 {
		return nil, PageNotFound
	}

	p := q[qi]
	if qi != len(q)-1 {
		q = append(q[:qi], q[qi+1:]...)
		q = append(q, p)
		pt.queue = q
	}

	return p, nil
}

func (pt *PageTable) put(p *Page) {
	q := pt.queue
	q = append(q, p)
	if pt.maxQueueSize < int64(len(q)) {
		q = q[1:]
	}
	pt.queue = q
}

func readPage(ra io.ReaderAt, pageSize, pageIndex int64) (*Page, error) {
	if pageIndex < 0 || pageSize < 1 {
		return nil, InvalidArg
	}

	buf := make([]byte, pageSize)
	n, err := ra.ReadAt(buf, pageIndex*pageSize)

	if err != nil && err != io.EOF {
		m := fmt.Sprintf("pagecache.readPage: io.ReadAt(buf(=%d bytes), pageIndex(=%d) * pageSize(=%d)) failed",
			len(buf), pageIndex, pageSize)
		return nil, errors.Wrap(err, m)
	}

	if err == nil && int64(n) != pageSize {
		m := fmt.Sprintf("pagecache.readPage: io.ReadAt(buf(=%d bytes), pageIndex(=%d) * pageSize(=%d)) failed: expect to read %d bytes but %d:",
			len(buf), pageIndex, pageSize, pageSize, n)
		return nil, errors.Wrap(err, m)
	}

	p := &Page{
		index: pageIndex,
		b:     buf[:n],
	}

	return p, nil
}

func (c *PageCache) page(pageIndex int64) (*Page, error) {
	p, err := c.pageTable.get(pageIndex)
	if err != nil && err != PageNotFound {
		return nil, err
	}

	if err == PageNotFound {
		p, err = readPage(c.ra, c.pageTable.pageSize, pageIndex)
		if err != nil {
			return nil, err
		}
		c.pageTable.put(p)
	}
	return p, nil
}

func (c *PageCache) ReadAt(b []byte, off int64) (int, error) {
	if off < 0 {
		return 0, InvalidArg
	}
	if len(b) == 0 {
		return 0, nil
	}

	fi := off / c.pageTable.pageSize
	li := (off + int64(len(b))) / c.pageTable.pageSize

	// read first page
	o := off % c.pageTable.pageSize
	p, err := c.page(fi)
	if err != nil {
		return 0, err
	}
	if int64(len(p.b)) < o {
		return 0, io.EOF
	}

	buf := b
	n := copy(buf, p.b[o:])
	buf = buf[n:]

	// read other pages
	for i := fi + 1; i <= li; i++ {
		p, err = c.page(i)
		if err != nil {
			return n, err
		}

		m := copy(buf, p.b)
		n = n + m
		buf = buf[m:]

		if int64(len(p.b)) < c.pageTable.pageSize && len(buf) != 0 {
			return n, io.EOF
		}
	}

	return n, nil
}

func NewReaderAtSize(ra io.ReaderAt, pageSize, pageTableSize int64) (io.ReaderAt, error) {
	if pageSize < 1 || pageTableSize < 2 {
		return nil, InvalidArg
	}
	pageTable := &PageTable{
		queue:        nil,
		pageSize:     pageSize,
		maxQueueSize: pageTableSize,
	}

	c := &PageCache{
		ra:        ra,
		pageTable: pageTable,
	}

	return c, nil
}
