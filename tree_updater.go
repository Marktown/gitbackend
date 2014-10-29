package gitbackend

import (
	"fmt"
	"github.com/libgit2/git2go"
	"strings"
)

type TreeUpdater struct {
	repo *git.Repository
}

func (this *TreeUpdater) Update(tree *git.Tree, path string, blobOid *git.Oid) (oid *git.Oid, err error) {
	treeBuilder, err := this.treeBuilder(tree)
	if err != nil {
		fmt.Println(err)
		return
	}
	return this.updateTree(treeBuilder, path, blobOid)
}

func (this *TreeUpdater) treeBuilder(tree *git.Tree) (treeBuilder *git.TreeBuilder, err error) {
	if tree == nil {
		treeBuilder, err = this.repo.TreeBuilder()
	} else {
		treeBuilder, err = this.repo.TreeBuilderFromTree(tree)
	}
	return
}

func (this *TreeUpdater) updateTree(treeBuilder *git.TreeBuilder, path string, blobOid *git.Oid) (oid *git.Oid, err error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return this.updateTreeBlob(treeBuilder, parts[0], blobOid)
	} else {
		return this.updateTreeTree(treeBuilder, parts[0], parts[1], blobOid)
	}
}

func (this *TreeUpdater) updateTreeBlob(treebuilder *git.TreeBuilder, basename string, blobOid *git.Oid) (oid *git.Oid, err error) {
	err = treebuilder.Insert(basename, blobOid, int(git.FilemodeBlob))
	if err != nil {
		fmt.Println(err)
		return
	}
	return treebuilder.Write()
}

func (this *TreeUpdater) updateTreeTree(treebuilder *git.TreeBuilder, basename string, childsPath string, blobOid *git.Oid) (oid *git.Oid, err error) {
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
