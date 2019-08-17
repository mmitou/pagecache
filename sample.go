package pagecache

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"pagecache/gcsreader"
)

type GCSEvent struct {
	Bucket         string `json:"bucket"`
	Name           string `json:"name"`
	Metageneration string `json:"metageneration"`
	ResourceState  string `json:"resourceState"`
}

var (
	storageClient *storage.Client
)

func init() {
	var err error

	storageClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}
}

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
