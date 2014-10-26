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

func NewFileStore(path string, isBare bool) (fileStore FileStore, err error) {
	repo, err := git.InitRepository(path, isBare)
	if err != nil {
		return
	}
	fileStore.repo = repo
	return
}

func (this *FileStore) ReadDir(path string) (list []FileInfo, err error) {
	if strings.Trim(path, "/ ") == "" {
		return this.readRootDir()
	} else {
		return this.readSubDir(path)
	}
}

func (this *FileStore) readRootDir() (list []FileInfo, err error) {
	headCommitTree, err, noHead := this.headCommitTree()
	if err != nil {
		// return empty list for newly initialized repository without proper HEAD
		// usually the first commit sets a proper HEAD
		// this is only necessary for the root directory since there are no files after init
		if noHead {
			err = nil
		}
		return
	}
	list = this.listTree(headCommitTree)
	return
}

func (this *FileStore) readSubDir(path string) (list []FileInfo, err error) {
	headCommitTree, err, _ := this.headCommitTree()
	if err != nil {
		return
	}

	entry, err := headCommitTree.EntryByPath(path)
	if err != nil {
		return
	}
	tree, err := this.repo.LookupTree(entry.Id)
	if err != nil {
		return
	}

	list = this.listTree(tree)
	return
}

func (this *FileStore) listTree(tree *git.Tree) (list []FileInfo) {
	var i uint64
	for i = 0; i < tree.EntryCount(); i++ {
		entry := tree.EntryByIndex(i)
		isDir := entry.Type == git.ObjectTree
		list = append(list, FileInfo{entry.Name, isDir})
	}
	return
}

func (this *FileStore) Checksum(path string) (hexdigest string, err error) {
	headCommitTree, err, _ := this.headCommitTree()
	if err != nil {
		fmt.Println(err)
		return
	}
	entry, err := headCommitTree.EntryByPath(path)
	if err != nil {
		return
	}
	hexdigest = entry.Id.String()
	return
}

func (this *FileStore) ReadFile(path string) (reader io.Reader, err error) {
	headCommitTree, err, _ := this.headCommitTree()
	if err != nil {
		fmt.Println(err)
		return
	}
	entry, err := headCommitTree.EntryByPath(path)
	if err != nil {
		return
	}
	blob, err := this.repo.LookupBlob(entry.Id)
	if err != nil {
		return
	}
	reader = bytes.NewBuffer(blob.Contents())
	return
}

func (this *FileStore) WriteFile(path string, reader io.Reader, commitInfo *CommitInfo) (err error) {
	blobOid, err := this.writeData(reader)

	if err != nil {
		fmt.Println(err)
		return
	}

	oldTree, _, _ := this.headCommitTree()
	newTreeId, err := this.updateTree(oldTree, path, blobOid)
	if err != nil {
		fmt.Println(err)
		return
	}

	tree, err := this.repo.LookupTree(newTreeId)
	if err != nil {
		fmt.Println(err)
		return
	}

	sig := &git.Signature{
		Name:  commitInfo.AuthorName(),
		Email: commitInfo.AuthorEmail(),
		When:  commitInfo.Time(),
	}

	commit, _, _ := this.headCommit()
	if commit == nil {
		_, err = this.repo.CreateCommit("HEAD", sig, sig, commitInfo.Message(), tree)

	} else {
		_, err = this.repo.CreateCommit("HEAD", sig, sig, commitInfo.Message(), tree, commit)

	}
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (this *FileStore) writeData(reader io.Reader) (blobOid *git.Oid, err error) {
	odb, err := this.repo.Odb()
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

func (this *FileStore) updateTree(oldParentTree *git.Tree, path string, blobOid *git.Oid) (oid *git.Oid, err error) {
	var treebuilder *git.TreeBuilder
	if oldParentTree == nil {
		treebuilder, err = this.repo.TreeBuilder()
	} else {
		treebuilder, err = this.repo.TreeBuilderFromTree(oldParentTree)
	}
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
	newTree, err := this.repo.LookupTree(newTreeOid)
	if err != nil {
		fmt.Println(err)
		return
	}

	oldChildTreeTreeEntry := newTree.EntryByName(parts[0])
	if oldChildTreeTreeEntry == nil {
		err = fmt.Errorf("Could not find Entry by Name %s", parts[0])
		return
	}
	oldChildTree, err2 := this.repo.LookupTree(oldChildTreeTreeEntry.Id)
	if err2 != nil {
		fmt.Println(err2)
		err = err2
		return
	}
	childTreeOid, err2 := this.updateTree(oldChildTree, parts[1], blobOid)
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

func (this *FileStore) headCommitTree() (tree *git.Tree, err error, noHead bool) {
	commit, err, noHead := this.headCommit()
	if err != nil {
		return
	}
	tree, err = commit.Tree()
	return
}

func (this *FileStore) headCommit() (commit *git.Commit, err error, noHead bool) {
	oid, err, noHead := this.headCommitId()
	if err != nil {
		return
	}
	commit, err = this.repo.LookupCommit(oid)
	return
}

func (this *FileStore) headCommitId() (oid *git.Oid, err error, noHead bool) {
	headRef, err := this.repo.LookupReference("HEAD")
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
