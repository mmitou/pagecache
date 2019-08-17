package pagecache

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestReadAt(t *testing.T) {
	bufsize := 100 * 1000 * 1000
	buffer := make([]byte, bufsize)
	rand.Read(buffer)
	r := bytes.NewReader(buffer)
	s, _ := NewReaderAtSize(r, 1024, 100)

	readSize := 256
	p := make([]byte, readSize)
	q := make([]byte, readSize)
	for i := 0; bufsize < i*readSize; i++ {
		m, err := r.ReadAt(p, int64(readSize*i))
		n, frr := s.ReadAt(q, int64(readSize*i))
		if m != n {
			t.Errorf("%d != %d", m, n)
		}
		if (err == nil && frr != nil) || (err != nil && frr == nil) {
			t.Errorf("%v != %v", err, frr)
		}
		if (err != nil && frr != nil) && (err.Error() != frr.Error()) {
			t.Errorf("%v != %v", err, frr)
		}

		for i := 0; i < m; i++ {
			if p[i] != q[i] {
				t.Errorf("[%d] %d != %d", i, p[i], q[i])
			}
		}
	}
}
