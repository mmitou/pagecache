pagecache
=========

- provides a cache layer that wraps object which also implements io.ReaderAt by page cache
- implements io.ReaderAt interface
- reads the wrapped underlayer by page
- manages the pages by page table
- drops a page by LRU(Least Recently Used) when the page table is full.
 
## usage 

Import pagecache, make an io.ReaderAt object wraps io.ReaderAt as below and call ReadAt.

```go
import (
	"io"
	"github.com/mmitou/pagecache"
)

func MyRead(readerAt io.ReaderAt, buf []byte) (int, error) {
	pageSize := int64(1024)
	maxNumOfPages := int64(10)
	off := int64(0)
	pcache := pagecache.NewReaderAtSize(readerAt, pageSize, maxNumOfPages)
	return p.ReadAt(buf, off)
}
```

## Example: unzip a big zip file in google cloud storage by google cloud functions and go.

### 1. checkout pagecache sample code

download sample.go.

```sh
git clone git@github.com:mmitou/pagecache.git
```

this Unzip function is defined in [pagecache/sample.go](https://github.com/mmitou/pagecache/blob/master/sample.go).

```go
func Unzip(ctx context.Context, e GCSEvent) error {
	o := storageClient.Bucket(e.Bucket).Object(e.Name)

	g, err := gcsreader.NewReader(ctx, o)
	if err != nil {
		return err
	}

	cra, err := NewReaderAtSize(g, 1048576, 10)
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(cra, g.Size())
	if err != nil {
		return err
	}

	baseDir := strings.TrimSuffix(e.Name, filepath.Ext(e.Name))
	buffer := make([]byte, 32*1024)
	outputBucket := os.Getenv("OUTPUTBUCKET")
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		err := func() error {
			r, err := f.Open()
			if err != nil {
				return fmt.Errorf("Open: %v", err)
			}
			defer r.Close()

			p := filepath.Join(baseDir, f.Name)
			w := storageClient.Bucket(outputBucket).Object(p).NewWriter(ctx)
			defer w.Close()

			_, err = io.CopyBuffer(w, r, buffer)
			if err != nil {
				return fmt.Errorf("io.Copy: %v", err)
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}
```

### 2. make buckets for input and output in google cloud storage 

```sh
gsutil mb gs://${INPUTBUCKET}
gsutil mb gs://${OUTPUTBUCKET}
```

### 3. deploy Unzip to google cloud functions

```sh
gcloud functions deploy unzip --set-env-vars OUTPUTBUCKET=${OUTPUT_BUCKET_ NAME}  --runtime go111 --entry-point Unzip --trigger-bucket=${INPUT_BUCKET_NAME} --region ${REGION_NAME} --source .
```

### 4. upload zip file you want to decompress.

```sh
gsutil cp zipfile.zip gs://${INPUT_BUCKET_NAME}
```

### 5. check out output

```sh
gsutil ls gs://${OUTPUT_BUCKET_NAME}
```
