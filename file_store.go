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
	head *Head
}

func NewFileStore(path string, isBare bool) (fileStore FileStore, err error) {
	repo, err := git.InitRepository(path, isBare)
	if err != nil {
		return
	}
	fileStore.repo = repo
	fileStore.head = &Head{repo}
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
	headCommitTree, err, noHead := this.head.CommitTree()
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
	headCommitTree, err, _ := this.head.CommitTree()
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
	headCommitTree, err, _ := this.head.CommitTree()
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
	headCommitTree, err, _ := this.head.CommitTree()
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

func (this *FileStore) CreateDir(path string, commitInfo *CommitInfo) (err error) {
	reader := strings.NewReader("")
	err = this.WriteFile(fmt.Sprintf("%s/.gitkeep", path), reader, commitInfo)
	return
}

func (this *FileStore) WriteFile(path string, reader io.Reader, commitInfo *CommitInfo) (err error) {
	blobOid, err := this.writeData(reader)

	if err != nil {
		fmt.Println(err)
		return
	}

	oldTree, _, _ := this.head.CommitTree()

	treeBuilder, err := this.treeBuilder(oldTree)
	if err != nil {
		fmt.Println(err)
		return
	}

	newTreeId, err := this.updateTree(treeBuilder, path, blobOid)
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

	commit, _, _ := this.head.Commit()
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

func (this *FileStore) treeBuilder(tree *git.Tree) (treeBuilder *git.TreeBuilder, err error) {
	if tree == nil {
		treeBuilder, err = this.repo.TreeBuilder()
	} else {
		treeBuilder, err = this.repo.TreeBuilderFromTree(tree)
	}
	return
}

func (this *FileStore) updateTree(treeBuilder *git.TreeBuilder, path string, blobOid *git.Oid) (oid *git.Oid, err error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return this.updateTreeBlob(treeBuilder, parts[0], blobOid)
	} else {
		return this.updateTreeTree(treeBuilder, parts[0], parts[1], blobOid)
	}
}

func (this *FileStore) updateTreeBlob(treebuilder *git.TreeBuilder, basename string, blobOid *git.Oid) (oid *git.Oid, err error) {
	err = treebuilder.Insert(basename, blobOid, int(git.FilemodeBlob))
	if err != nil {
		fmt.Println(err)
		return
	}
	return treebuilder.Write()
}

func (this *FileStore) updateTreeTree(treebuilder *git.TreeBuilder, basename string, childsPath string, blobOid *git.Oid) (oid *git.Oid, err error) {
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

	var oldChildTree *git.Tree

	// try to fetch an existing tree entry
	// if no tree entry is found, a new one will automatically created later
	oldChildTreeTreeEntry := newTree.EntryByName(basename)
	if oldChildTreeTreeEntry != nil {
		oldChildTree, err = this.repo.LookupTree(oldChildTreeTreeEntry.Id)
	}

	childTreeBuilder, err := this.treeBuilder(oldChildTree)
	if err != nil {
		fmt.Println(err)
		return
	}

	childTreeOid, err2 := this.updateTree(childTreeBuilder, childsPath, blobOid)
	if err2 != nil {
		fmt.Println(err2)
		err = err2
		return
	}

	err = treebuilder.Insert(basename, childTreeOid, int(git.FilemodeTree))
	if err != nil {
		fmt.Println(err)
		return
	}

	oid, err = treebuilder.Write()
	return
}
