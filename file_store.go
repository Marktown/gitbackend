package gitbackend

import (
	"bytes"
	"fmt"
	"github.com/libgit2/git2go"
	"io"
	"io/ioutil"
	"strings"
)

type FileStore struct {
  repo *git.Repository
}

func NewFileStore(path string, isBare bool) (fs FileStore, err error) {
	repo, err := git.InitRepository(path, isBare)
	if err != nil {
		return
	}
	fs.repo = repo
	return
}

func (f *FileStore) ReadDir(path string) (list []FileInfo, err error) {
	if strings.Trim(path, "/ ") == "" {
		return f.readRootDir()
	} else {
		return f.readSubDir(path)
	}
}

func (f *FileStore) readRootDir() (list []FileInfo, err error) {
	headCommitTree, err, noHead := f.headCommitTree()
	if err != nil {
		// return empty list for newly initialized repository without proper HEAD
		// usually the first commit sets a proper HEAD
		// this is only necessary for the root directory since there are no files after init
		if noHead {
			err = nil
		}
		return
	}
	list = f.listTree(headCommitTree)
	return
}

func (f *FileStore) readSubDir(path string) (list []FileInfo, err error) {
	headCommitTree, err, _ := f.headCommitTree()
	if err != nil {
		return
	}

	entry, err := headCommitTree.EntryByPath(path)
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
		isDir := entry.Type == git.ObjectTree
		list = append(list, FileInfo{entry.Name, isDir})
	}
	return
}

func (f *FileStore) ReadFile(path string) (reader io.Reader, err error) {
	headCommitTree, err, _ := f.headCommitTree()
	if err != nil {
		fmt.Println(err)
		return
	}
	entry, err := headCommitTree.EntryByPath(path)
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
	blobOid, err := f.writeData(reader)

	if err != nil {
		fmt.Println(err)
		return
	}

	oldTree, err, _ := f.headCommitTree()
	if err != nil {
		fmt.Println(err)
		return
	}

	newTreeId, err := f.updateTree(oldTree, path, blobOid)
	if err != nil {
		fmt.Println(err)
		return
	}

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
		fmt.Println(err)
		return
	}
	return
}

func (f *FileStore) writeData(reader io.Reader) (blobOid *git.Oid, err error) {
	odb, err := f.repo.Odb()
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(err)
		return
	}

	blobOid, err = odb.Write(data, git.ObjectBlob)
	return
}

func (f *FileStore) updateTree(oldParentTree *git.Tree, path string, blobOid *git.Oid) (oid *git.Oid, err error) {
	treebuilder, err := f.repo.TreeBuilderFromTree(oldParentTree)
	if err != nil {
		fmt.Println(err)
		return
	}
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		err = treebuilder.Insert(parts[0], blobOid, int(git.FilemodeBlob))
		if err != nil {
			fmt.Println(err)
			return
		}
		return treebuilder.Write()
	}

	newTreeOid, err := treebuilder.Write()
	if err != nil {
		fmt.Println(err)
		return
	}
	newTree, err := f.repo.LookupTree(newTreeOid)
	if err != nil {
		fmt.Println(err)
		return
	}

	oldChildTreeTreeEntry := newTree.EntryByName(parts[0])
	if oldChildTreeTreeEntry == nil {
		err = fmt.Errorf("Could not find Entry by Name %s", parts[0])
		return
	}
	oldChildTree, err2 := f.repo.LookupTree(oldChildTreeTreeEntry.Id)
	if err2 != nil {
		fmt.Println(err2)
		err = err2
		return
	}
	childTreeOid, err2 := f.updateTree(oldChildTree, parts[1], blobOid)
	if err2 != nil {
		fmt.Println(err2)
		err = err2
		return
	}
	err = treebuilder.Insert(parts[0], childTreeOid, int(git.FilemodeTree))
	if err != nil {
		fmt.Println(err)
		return
	}

	oid, err = treebuilder.Write()
	return
}

func (f *FileStore) headCommitTree() (tree *git.Tree, err error, noHead bool) {
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
