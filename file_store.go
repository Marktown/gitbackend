package gitbackend

import (
	"bytes"
	"fmt"
	"github.com/libgit2/git2go"
	"io"
	"time"
)

type CommitInfo struct {
	authorName, authorEmail, message string
	time                             time.Time
}

func (c *CommitInfo) AuthorName() string {
	return c.authorName
}

func (c *CommitInfo) AuthorEmail() string {
	return c.authorEmail
}

func (c *CommitInfo) Time() time.Time {
	return c.time
}

func (c *CommitInfo) Message() string {
	return c.message
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

type FileStore struct {
	repo *git.Repository
}

func (f *FileStore) ReadRoot() (list []FileInfo, err error) {
	tree, err, noHead := f.tree()
	if err != nil {
		// return empty list for newly initialized repository without proper HEAD
		// usually the first commit sets a proper HEAD
		// this is only necessary for the root directory since there are no files after init
		if noHead {
			err = nil
		}
		return
	}
	list = f.listTree(tree)
	return
}

func (f *FileStore) ReadDir(path string) (list []FileInfo, err error) {
	root, err, _ := f.tree()
	if err != nil {
		return
	}

	entry, err := root.EntryByPath(path)
	if err != nil {
		return
	}
	tree, err := f.repo.LookupTree(entry.Id)
	if err != nil {
		return
	}

	list = f.listTree(tree)
	return
}

func (f *FileStore) listTree(tree *git.Tree) (list []FileInfo) {
	var i uint64
	for i = 0; i < tree.EntryCount(); i++ {
		entry := tree.EntryByIndex(i)
		list = append(list, FileInfo{entry.Name})
	}
	return
}

func (f *FileStore) ReadFile(path string) (reader io.Reader, err error) {
	tree, err, _ := f.tree()
	if err != nil {
		fmt.Println(err)
		return
	}
	entry, err := tree.EntryByPath(path)
	if err != nil {
		return
	}
	blob, err := f.repo.LookupBlob(entry.Id)
	if err != nil {
		return
	}
	reader = bytes.NewBuffer(blob.Contents())
	return
}

func (f *FileStore) WriteFile(path string, reader io.Reader, commitInfo CommitInfo) (err error) {
	odb, err := f.repo.Odb()
	if err != nil {
		fmt.Println(err)
		return
	}

	blobOid, err := odb.Write([]byte(readAll(reader)), git.ObjectBlob)
	if err != nil {
		fmt.Println(err)
		return
	}

	oldTree, err, _ := f.tree()
	if err != nil {
		fmt.Println(err)
		return
	}
	treebuilder, err := f.repo.TreeBuilderFromTree(oldTree)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = treebuilder.Insert(path, blobOid, git.FilemodeBlob)
	if err != nil {
		fmt.Println(err)
		return
	}
	newTreeId, err := treebuilder.Write()
	tree, err := f.repo.LookupTree(newTreeId)
	if err != nil {
		fmt.Println(err)
		return
	}

	sig := &git.Signature{
		Name:  commitInfo.AuthorName(),
		Email: commitInfo.AuthorEmail(),
		When:  commitInfo.Time(),
	}

	commit, err, _ := f.headCommit()
	if err != nil {
		return
	}

	_, err = f.repo.CreateCommit("HEAD", sig, sig, commitInfo.Message(), tree, commit)
	if err != nil {
		fmt.Println("1")
		fmt.Println(err)
		return
	}
	return
}

func (f *FileStore) tree() (tree *git.Tree, err error, noHead bool) {
	commit, err, noHead := f.headCommit()
	if err != nil {
		return
	}
	tree, err = commit.Tree()
	return
}

func (f *FileStore) headCommit() (commit *git.Commit, err error, noHead bool) {
	oid, err, noHead := f.headCommitId()
	if err != nil {
		return
	}
	commit, err = f.repo.LookupCommit(oid)
	return
}

func (f *FileStore) headCommitId() (oid *git.Oid, err error, noHead bool) {
	headRef, err := f.repo.LookupReference("HEAD")
	if err != nil {
		return
	}
	ref, err := headRef.Resolve()
	if err != nil {
		noHead = true
		return
	}
	oid = ref.Target()
	if oid == nil {
		err = fmt.Errorf("Could not get Target for HEAD(%s)\n", oid.String())
	}
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
