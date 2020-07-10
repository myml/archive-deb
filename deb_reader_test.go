package deb

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestReaderAll(t *testing.T) {
	for _, filename := range testFiles() {
		testReader(filename)
	}
}

func TestReader(t *testing.T) {
	err := testReader(filepath.Join(testdata, testfile))
	if err != nil {
		t.Fatal(err)
	}
}

func testReader(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	r := NewReader(f)
	for {
		head, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if head.Size == 0 {
			continue
		}
		n, err := io.Copy(ioutil.Discard, r)
		if err != nil {
			return err
		}
		if head.Size != n {
			return io.ErrShortWrite
		}
	}
	return nil
}
