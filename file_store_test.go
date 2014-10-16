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

	paths, err := fileStore.ReadRoot()
	if err != nil {
		T.Fatal(err)
	}
	if len(paths) != 0 {
		T.Fatalf("paths should have length 0, but is %d", len(paths))
	}

	_, err = fileStore.ReadDir("foo")
	if err == nil {
		T.Fatal("expected error, but nothing was returned")
	}
}

func TestReadRoot(T *testing.T) {
	path := "tests/repo"
	fileStore, err := NewFileStore(path, false)
	if err != nil {
		T.Fatal(err)
	}

	paths, err := fileStore.ReadRoot()
	if err != nil {
		T.Fatal(err)
	}
	if len(paths) != 2 {
		T.Fatalf("paths should have length 1, but is %d", len(paths))
	}

	if paths[0].Name() != "bar" {
		T.Fatalf("First path should be bar, but is %s", paths[0].Name())
	}
}

func TestReadDir(T *testing.T) {
	path := "tests/repo"
	fileStore, err := NewFileStore(path, false)
	if err != nil {
		T.Fatal(err)
	}

	paths, err := fileStore.ReadDir("bar")
	if err != nil {
		T.Fatal(err)
	}
	if len(paths) != 1 {
		T.Fatalf("paths should have length 1, but is %d", len(paths))
	}

	if paths[0].Name() != "baz.txt" {
		T.Fatalf("First path should be foo.txt, but is %s", paths[0].Name())
	}

	_, err = fileStore.ReadDir("foo")
	if err == nil {
		T.Fatal("expected error, but nothing was returned")
	}
}

func TestReadFile(T *testing.T) {
	path := "tests/repo"
	fileStore, err := NewFileStore(path, false)
	if err != nil {
		T.Fatal(err)
	}

	reader, err := fileStore.ReadFile("foo.txt")
	if err != nil {
		T.Fatal(err)
	}
	s := readAll(reader)
	if s != fmt.Sprintf("Hello World\n") {
		T.Fatalf("Expected: 'Hello World\n'\nactual: '%s'", s)
	}

	reader, err = fileStore.ReadFile("bar/baz.txt")
	if err != nil {
		T.Fatal(err)
	}
	s = readAll(reader)
	if s != fmt.Sprintf("This is Baz\n") {
		T.Fatalf("Expected: 'This is Baz\n'\nactual: '%s'", s)
	}

	reader, err = fileStore.ReadFile("boo.txt")
	if err == nil {
		T.Fatal("expected error, but nothing was returned")
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
