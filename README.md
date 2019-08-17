pagecache
=========

pagecache is a cache layer that wraps io.ReaderAt by page cache.<br />
pagecache implements io.ReaderAt interface, reads the wrapped underlayer by page, and manages the pages by page table.<br />
to specify proper page size and number of pages, a program which has to run in severe memory environment may be able to read a file sizes over memory.

## usage 

Import pagecache.

```go
import github.com/mmitou/pagecache
```

## example 

```go
import (
	"io"
	"github.com/mmitou/pagecache"
)

func MyRead(readerAt io.ReaderAt, buf []byte) (int, error) {
	pageSize := int64(1024)
	numOfPage := int64(10)
	off := int64(0)
	pcache := pagecache.NewReaderAtSize(readerAt, pageSize, numOfPage)
	return p.ReadAt(buf, off)
}
```
