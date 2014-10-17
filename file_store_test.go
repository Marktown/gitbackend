package gitbackend

import (
	"fmt"
	"github.com/libgit2/git2go"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"text/scanner"
	"time"
)

func TestNewFS(t *testing.T) {
	path := "tmp/bla/test"
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
	fileStore, err := NewFileStore(path, true)
	if err != nil {
		t.Fatal(err)
	}
	fileInfo, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !fileInfo.IsDir() {
		t.Fatalf("%s is not a directory.", path)
	}
	repo, err := git.OpenRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.IsBare() {
		t.Fatalf("%s is not a Bare Repository.", path)
	}

	paths, err := fileStore.ReadRoot()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 0 {
		t.Fatalf("paths should have length 0, but is %d", len(paths))
	}

	_, err = fileStore.ReadDir("foo")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}
}

func TestReadRoot(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	fileStore, err := NewFileStore(repo.Workdir(), false)
	if err != nil {
		t.Fatal(err)
	}

	paths, err := fileStore.ReadRoot()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("paths should have length 1, but is %d", len(paths))
	}

	if paths[0].Name() != "bar" {
		t.Fatalf("First path should be bar, but is %s", paths[0].Name())
	}
}

func TestReadDir(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	fileStore, err := NewFileStore(repo.Workdir(), false)
	if err != nil {
		t.Fatal(err)
	}

	paths, err := fileStore.ReadDir("bar")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("paths should have length 1, but is %d", len(paths))
	}

	if paths[0].Name() != "baz.txt" {
		t.Fatalf("First path should be foo.txt, but is %s", paths[0].Name())
	}

	_, err = fileStore.ReadDir("foo")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}
}

func TestReadFile(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	fileStore, err := NewFileStore(repo.Workdir(), false)
	if err != nil {
		t.Fatal(err)
	}

	reader, err := fileStore.ReadFile("foo.txt")
	if err != nil {
		t.Fatal(err)
	}
	s := readAll(reader)
	if s != fmt.Sprintf("Hello World\n") {
		t.Fatalf("Expected: 'Hello World\n'\nactual: '%s'", s)
	}

	reader, err = fileStore.ReadFile("bar/baz.txt")
	if err != nil {
		t.Fatal(err)
	}
	s = readAll(reader)
	if s != fmt.Sprintf("This is Baz\n") {
		t.Fatalf("Expected: 'This is Baz\n'\nactual: '%s'", s)
	}

	reader, err = fileStore.ReadFile("boo.txt")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}
}

func createTestRepo(t *testing.T) *git.Repository {
	// figure out where we can create the test repo
	path, err := ioutil.TempDir("", "test_repo")
	checkFatal(t, err)
	repo, err := git.InitRepository(path, false)
	checkFatal(t, err)

	return repo
}

func seedTestRepo(t *testing.T, repo *git.Repository) (*git.Oid, *git.Oid) {
	err := exec.Command("cp", "-rf", "tests/repo/", repo.Workdir()).Run()
	checkFatal(t, err)

	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("bar/baz.txt")
	checkFatal(t, err)
	err = idx.AddByPath("foo.txt")
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	return commitId, treeId
}

func checkFatal(t *testing.T, err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatal()
	}

	t.Fatalf("Fail at %v:%v; %v", file, line, err)
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
