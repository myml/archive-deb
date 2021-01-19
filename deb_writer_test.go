package deb

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWriterAll(t *testing.T) {
	for _, filename := range testFiles() {
		t.Log(filename)
		testWriter(t, filename)
	}
}

func TestWriter(t *testing.T) {
	err := testWriter(t, filepath.Join(testdata, testfile))
	if err != nil {
		t.Fatal(err)
	}
}

func testWriter(t *testing.T, filename string) error {
	rf, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer rf.Close()

	wf, err := ioutil.TempFile("", "tmp_")
	if err != nil {
		return err
	}
	defer os.Remove(wf.Name())
	defer wf.Close()

	r := NewReader(rf)
	w := NewWriter(wf)
	for {
		head, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		err = w.WriteHeader(head)
		if err != nil {
			return err
		}
		if head.Size == 0 {
			continue
		}
		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("deb close: %w", err)
	}
	err = wf.Close()
	if err != nil {
		return fmt.Errorf("file sync: %w", err)
	}
	err = exec.Command("ar", "-t", wf.Name()).Run()
	if err != nil {
		return fmt.Errorf("ar command verify: %w", err)
	}
	err = exec.Command("dpkg-deb", "--info", wf.Name()).Run()
	if err != nil {
		return fmt.Errorf("dpkg command verify: %w", err)
	}
	return nil
}
