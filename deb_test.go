package deb

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ATestReader(t *testing.T) {
	for _, filename := range testFiles() {
		testReader(filename)
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
		if err == io.EOF {
			break
		}
		fmt.Println(filename, head.Name)
		if head.FileInfo().IsDir() {
			continue
		}
		n, err := io.Copy(ioutil.Discard, r)
		if err != nil {
			return err
		}
		if head.Size != n {
			return fmt.Errorf("read file size %d != %d", head.Size, n)
		}
	}
	return nil
}

func TestWriter(t *testing.T) {
	for _, filename := range testFiles() {
		if strings.HasSuffix(filename, ".test.deb") {
			continue
		}
		err := testWriter(filename)
		if err != nil {
			t.Fatal(err)
		}
	}
}
func testWriter(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	wf, err := os.Create(filename + ".test.deb")
	if err != nil {
		return err
	}
	defer wf.Close()
	r := NewReader(f)
	w := NewWriter(wf)
	defer w.Close()
	for {
		head, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = w.WriteHeader(head)
		if err != nil {
			return err
		}
		if head.FileInfo().IsDir() {
			continue
		}
		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
		log.Println(filename, head.Name)
	}
	return nil
}

func testFiles() []string {
	list, err := filepath.Glob("/home/myml/Src/dpkg-archiver/testdata/*")
	if err != nil {
		panic(err)
	}
	return list
}
