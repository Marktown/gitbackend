package gitbackend

import (
	"fmt"
	"github.com/libgit2/git2go"
	"io"
	"os"
	"testing"
	"text/scanner"
)

func TestNewFS(T *testing.T) {
	path := "tmp/bla/test"
	err := os.RemoveAll(path)
	if err != nil {
		T.Fatal(err)
	}
	fileStore, err := NewFileStore(path, true)
	if err != nil {
		T.Fatal(err)
	}
	fileInfo, err := os.Lstat(path)
	if err != nil {
		T.Fatal(err)
	}
	if !fileInfo.IsDir() {
		T.Fatalf("%s is not a directory.", path)
	}
	repo, err := git.OpenRepository(path)
	if err != nil {
		T.Fatal(err)
	}
	if !repo.IsBare() {
		T.Fatalf("%s is not a Bare Repository.", path)
	}

	paths, err := fileStore.ReadDir("/")
	if err != nil {
		T.Fatal(err)
	}
	if len(paths) != 0 {
		T.Fatalf("paths should have length 0, but is %d", len(paths))
	}
}

func TestReadDir(T *testing.T) {
	path := "test/repo"
	fileStore, err := NewFileStore(path, false)
	if err != nil {
		T.Fatal(err)
	}

	paths, err := fileStore.ReadDir("/")
	if err != nil {
		T.Fatal(err)
	}
	if len(paths) != 1 {
		T.Fatalf("paths should have length 1, but is %d", len(paths))
	}

	if paths[0].Name() != "foo.txt" {
		T.Fatalf("First path should be foo.txt, but is %s", paths[0].Name())
	}
}

func TestReadFile(T *testing.T) {
	path := "test/repo"
	fileStore, err := NewFileStore(path, false)
	if err != nil {
		T.Fatal(err)
	}

	reader, err := fileStore.ReadFile("foo.txt")
	if err != nil {
		T.Fatal(err)
	}
	s := readAll(reader)
	if s != fmt.Sprintf("Hello World\n\n") {
		T.Fatalf("Expected: 'Hello World\\n'\nactual: '%s'", s)
	}
}

func readAll(reader io.Reader) (data string) {
	data = ""
	var s scanner.Scanner
	s.Init(reader)
	s.Whitespace = 1
	tok := s.Scan()
	for tok != scanner.EOF {
		data += s.TokenText()
		tok = s.Scan()
	}
	return
}
