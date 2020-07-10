package deb

import (
	"os"
	"path/filepath"
	"testing"
)

const testdata = "./testdata"

var testfile = "test.deb"

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func testFiles() []string {
	list, err := filepath.Glob("testdata/*")
	if err != nil {
		panic(err)
	}
	return list
}
