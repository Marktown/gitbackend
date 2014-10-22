package gitbackend

import (
	"bufio"
	"fmt"
	"github.com/libgit2/git2go"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNewFS(t *testing.T) {
	path := "tmp/bla/test"
	err := os.RemoveAll(path)
	checkFatal(t, err)
	fileStore, err := NewFileStore(path, true)
	checkFatal(t, err)
	fileInfo, err := os.Lstat(path)
	checkFatal(t, err)
	if !fileInfo.IsDir() {
		t.Fatalf("%s is not a directory.", path)
	}
	repo, err := git.OpenRepository(path)
	checkFatal(t, err)
	if !repo.IsBare() {
		t.Fatalf("%s is not a Bare Repository.", path)
	}

	paths, err := fileStore.ReadDir("/")
	checkFatal(t, err)
	if len(paths) != 0 {
		t.Fatalf("paths should have length 0, but is %d", len(paths))
	}

	paths, err = fileStore.ReadDir("foo")
	if err == nil {
		t.Fatal("expected error, but nothing was returned. instead %v", paths)
	}
}

func TestReadDirRoot(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	fileStore, err := NewFileStore(repo.Workdir(), false)
	checkFatal(t, err)

	paths, err := fileStore.ReadDir("/")
	checkFatal(t, err)
	if len(paths) != 2 {
		t.Fatalf("paths should have length 2, but is %d\npaths contains: %v", len(paths), paths)
	}

	if paths[0].Name() != "bar" {
		t.Fatalf("First path should be bar, but is %s\npaths contains: %v", paths[0].Name(), paths)
	}
	if !paths[0].IsDir() {
		t.Fatalf("First path should IsDir() but is not %v", paths)
	}
}

func TestReadDir(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	fileStore, err := NewFileStore(repo.Workdir(), false)
	checkFatal(t, err)

	paths, err := fileStore.ReadDir("bar")
	checkFatal(t, err)
	if len(paths) != 1 {
		t.Fatalf("paths should have length 1, but is %d\npaths contains: %v", len(paths), paths)
	}

	if paths[0].Name() != "baz.txt" {
		t.Fatalf("First path should be foo.txt, but is %s\npaths contains: %v", paths[0].Name(), paths)
	}

	if paths[0].IsDir() {
		t.Fatalf("First path should not IsDir() but is", paths)
	}

	_, err = fileStore.ReadDir("foo")
	if err == nil {
		t.Fatalf("expected error, but nothing was returned\npaths contains: %v", paths)
	}
}

func TestChecksum(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	fileStore, err := NewFileStore(repo.Workdir(), false)
	checkFatal(t, err)

	hexdigest, err := fileStore.Checksum("foo.txt")
	checkFatal(t, err)
	if hexdigest != "557db03de997c86a4a028e1ebd3a1ceb225be238" {
		t.Fatalf("Expected: 557db03de997c86a4a028e1ebd3a1ceb225be238\nactual: '%s'", hexdigest)
	}

	hexdigest, err = fileStore.Checksum("bar")
	checkFatal(t, err)
	if hexdigest != "c618c0d21bee6744d0800dc8c56bed7df3eb268c" {
		t.Fatalf("Expected: c618c0d21bee6744d0800dc8c56bed7df3eb268c\nactual: '%s'", hexdigest)
	}

	hexdigest, err = fileStore.Checksum("bar/baz.txt")
	checkFatal(t, err)
	if hexdigest != "988d3d574a8dcedd83d8d63868822f35ed789e5b" {
		t.Fatalf("Expected: 988d3d574a8dcedd83d8d63868822f35ed789e5b\nactual: '%s'", hexdigest)
	}
}

func TestReadFile(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	fileStore, err := NewFileStore(repo.Workdir(), false)
	checkFatal(t, err)

	reader, err := fileStore.ReadFile("foo.txt")
	checkFatal(t, err)
	data, _ := ioutil.ReadAll(reader)
	s := string(data)

	if s != fmt.Sprintf("Hello World\n") {
		t.Fatalf("Expected: 'Hello World\n'\nactual: '%s'", s)
	}

	reader, err = fileStore.ReadFile("bar/baz.txt")
	checkFatal(t, err)

	data, _ = ioutil.ReadAll(reader)
	s = string(data)

	if s != fmt.Sprintf("This is Baz\n") {
		t.Fatalf("Expected: 'This is Baz\n'\nactual: '%s'", s)
	}

	reader, err = fileStore.ReadFile("boo.txt")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}
}

func TestWriteFile(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	fileStore, err := NewFileStore(repo.Workdir(), false)
	checkFatal(t, err)

	reader := strings.NewReader("Hello World")
	commitInfo := CommitInfo{"Paul", "p@example.com", "Have fun.", time.Date(2014, 10, 17, 13, 37, 0, 0, &time.Location{})}
	err = fileStore.WriteFile("bar.txt", reader, commitInfo)
	checkFatal(t, err)

	cmd := exec.Command("git", "show")
	cmd.Dir = repo.Workdir()
	b, err := cmd.Output()
	checkFatal(t, err)

	r := strings.NewReader(string(b))
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if l := scanner.Text(); !regexp.MustCompile("commit [a-f0-9]+").MatchString(l) {
		t.Fatalf("expected: commit ...\nactual: %s", l)
	}

	assertNextLineEqual("Author: Paul <p@example.com>", scanner, t)
	assertNextLineEqual("Date:   Fri Oct 17 13:37:00 2014 +0000", scanner, t)
	assertNextLineEqual("", scanner, t)
	assertNextLineEqual("    Have fun.", scanner, t)
	assertNextLineEqual("", scanner, t)
	assertNextLineEqual("diff --git a/bar.txt b/bar.txt", scanner, t)
	scanner.Scan()
	scanner.Scan()
	scanner.Scan()
	assertNextLineEqual("+++ b/bar.txt", scanner, t)
	scanner.Scan()
	assertNextLineEqual("+Hello World", scanner, t)
}

func TestWriteFileWithSubdir(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	fileStore, err := NewFileStore(repo.Workdir(), false)
	checkFatal(t, err)

	reader := strings.NewReader("Hello World")
	commitInfo := CommitInfo{"Paul", "p@example.com", "Have fun.", time.Date(2014, 10, 17, 13, 37, 0, 0, &time.Location{})}
	err = fileStore.WriteFile("bar/boo.txt", reader, commitInfo)
	checkFatal(t, err)

	cmd := exec.Command("git", "show")
	cmd.Dir = repo.Workdir()
	b, err := cmd.Output()
	checkFatal(t, err)

	r := strings.NewReader(string(b))
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if l := scanner.Text(); !regexp.MustCompile("commit [a-f0-9]+").MatchString(l) {
		t.Fatalf("expected: commit ...\nactual: %s", l)
	}

	assertNextLineEqual("Author: Paul <p@example.com>", scanner, t)
	assertNextLineEqual("Date:   Fri Oct 17 13:37:00 2014 +0000", scanner, t)
	assertNextLineEqual("", scanner, t)
	assertNextLineEqual("    Have fun.", scanner, t)
	assertNextLineEqual("", scanner, t)
	assertNextLineEqual("diff --git a/bar/boo.txt b/bar/boo.txt", scanner, t)
	scanner.Scan()
	scanner.Scan()
	scanner.Scan()
	assertNextLineEqual("+++ b/bar/boo.txt", scanner, t)
	scanner.Scan()
	assertNextLineEqual("+Hello World", scanner, t)
}

func assertNextLineEqual(expected string, scanner *bufio.Scanner, t *testing.T) {
	scanner.Scan()
	assertStringEqual(expected, scanner.Text(), t)
}

func assertStringEqual(expected, actual string, t *testing.T) {
	if actual != expected {
		t.Errorf("expected: %s\nactual: %s", expected, actual)
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
	err := exec.Command("cp", "-Rf", "tests/repo/.", repo.Workdir()).Run()
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
	filepath.Walk(repo.Workdir(), func(path string, info os.FileInfo, _ error) (err error) {
		if info.IsDir() {
			return
		}
		lenWorkdir := len(repo.Workdir())
		if path[lenWorkdir:lenWorkdir+4] == ".git" {
			return
		}
		err = idx.AddByPath(path[lenWorkdir:])
		checkFatal(t, err)
		return
	})
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "Initial commit.\n"
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
