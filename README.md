pagecache
=========

- provides a cache layer that wraps object which also implements io.ReaderAt by page cache
- implements io.ReaderAt interface
- reads the wrapped underlayer by page
- manages the pages by page table
- drops a page by LRU(Least Recently Used) when the page table is full.
 
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
