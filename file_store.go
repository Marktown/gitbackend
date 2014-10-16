package gitbackend

import (
	"bytes"
	"fmt"
	"github.com/libgit2/git2go"
	"io"
)

type FileStore struct {
	repo *git.Repository
}

type FileInfo struct {
	name string // base name of the file
	// path  string
	// size  int64 // length in bytes for regular files; system-dependent for others
	// isDir bool  // abbreviation for Mode().IsDir()
}

func (f *FileInfo) Name() string {
	return f.name
}

func (f *FileStore) ReadDir(path string) (list []FileInfo, err error) {
	headRef, err := f.repo.LookupReference("HEAD")
	if err != nil {
		fmt.Printf("Could not lookup HEAD: %v\n", err)
		return
	}

	ref, err := headRef.Resolve()
	if err != nil {
		fmt.Printf("Could not resolve HEAD: %v\n", err)
		return list, nil
	}

	oid := ref.Target()
	if oid == nil {
		err = fmt.Errorf("Could not get Target for HEAD(%s)\n", oid.String())
		fmt.Print(err)
		return
	}
	commit, err := f.repo.LookupCommit(oid)
	if err != nil {
		fmt.Printf("Could not lookup HEAD commit(%s): %v\n", oid.String(), err)
		return
	}

	tree, err := commit.Tree()
	if err != nil {
		fmt.Printf("Could not get Tree for HEAD commit(%s): %v\n", oid.String(), err)
		return
	}
	var i uint64
	for i = 0; i < tree.EntryCount(); i++ {
		entry := tree.EntryByIndex(i)
		list = append(list, FileInfo{entry.Name})
	}
	return
}

func (f *FileStore) ReadFile(path string) (reader io.Reader, err error) {
	headRef, err := f.repo.LookupReference("HEAD")
	if err != nil {
		fmt.Printf("Could not lookup HEAD: %v\n", err)
		return
	}

	ref, err := headRef.Resolve()
	if err != nil {
		fmt.Printf("Could not resolve HEAD: %v\n", err)
		return
	}

	oid := ref.Target()
	if oid == nil {
		err = fmt.Errorf("Could not get Target for HEAD(%s)\n", oid.String())
		fmt.Print(err)
		return
	}
	commit, err := f.repo.LookupCommit(oid)
	if err != nil {
		fmt.Printf("Could not lookup HEAD commit(%s): %v\n", oid.String(), err)
		return
	}

	tree, err := commit.Tree()
	if err != nil {
		fmt.Printf("Could not get Tree for HEAD commit(%s): %v\n", oid.String(), err)
		return
	}
	entry := tree.EntryByName(path)
	blob, err := f.repo.LookupBlob(entry.Id)
	if err != nil {
		fmt.Printf("Could not lookup Blob (%s): %v\n", entry.Id, err)
		return
	}
	reader = bytes.NewBuffer(blob.Contents())

	return
}

func NewFileStore(path string, isBare bool) (fs FileStore, err error) {
	repo, err := git.InitRepository(path, isBare)
	if err != nil {
		return
	}
	fs.repo = repo
	return
}
