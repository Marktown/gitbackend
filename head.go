package gitbackend

import (
	"fmt"
	"github.com/libgit2/git2go"
)

type Head struct {
	repo *git.Repository
}

func (this *Head) CommitTree() (tree *git.Tree, err error, noHead bool) {
	commit, err, noHead := this.Commit()
	if err != nil {
		return
	}
	tree, err = commit.Tree()
	return
}

func (this *Head) Commit() (commit *git.Commit, err error, noHead bool) {
	oid, err, noHead := this.CommitId()
	if err != nil {
		return
	}
	commit, err = this.repo.LookupCommit(oid)
	return
}

func (this *Head) CommitId() (oid *git.Oid, err error, noHead bool) {
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
