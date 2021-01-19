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
		testReader(t, filename)
	}
}

func TestReader(t *testing.T) {
	err := testReader(t, filepath.Join(testdata, testfile))
	if err != nil {
		t.Fatal(err)
	}
}

func testReader(t *testing.T, filename string) error {
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
		t.Log(head)
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
